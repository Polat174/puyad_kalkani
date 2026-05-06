package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/puyad/kalkani/internal/backup"
	"github.com/puyad/kalkani/internal/config"
	"github.com/puyad/kalkani/internal/executor"
	"github.com/puyad/kalkani/internal/sse"
	"github.com/puyad/kalkani/internal/sshguard"
)

// Server, API sunucusunu tutar.
type Server struct {
	Broker      *sse.Broker
	Baseline    *config.BaselineConfig
	Blacklist   *config.BlacklistConfig
	mu          sync.Mutex
	lastResults []executor.ScanResult
	lastReport  *executor.ScanReport
}

// NewServer, yeni bir API sunucusu oluşturur.
func NewServer(baseline *config.BaselineConfig, blacklist *config.BlacklistConfig) *Server {
	return &Server{
		Broker:    sse.NewBroker(),
		Baseline:  baseline,
		Blacklist: blacklist,
	}
}

// NewRouter, ileride HTTP uç noktaları için chi yönlendiricisi döndürür.
func NewRouter(srv *Server) *chi.Mux {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// API endpoints
	r.Get("/api/scan", srv.handleScan)
	r.Get("/api/results", srv.handleResults)
	r.Post("/api/fix/{id}", srv.handleFix)
	r.Post("/api/fix-all", srv.handleFixAll)
	r.Get("/api/backups", srv.handleBackups)
	r.Post("/api/restore/{id}", srv.handleRestore)

	// SSE stream
	r.Get("/api/events", srv.Broker.ServeHTTP)

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hostname, _ := os.Hostname()

	// Tarayıcıları oluştur
	scanners := []executor.Scanner{
		sshguard.NewSSHScanner(s.Baseline.SSH),
		executor.NewSysctlScanner(s.Baseline.Kernel),
		executor.NewFirewallScanner(s.Baseline.Firewall),
		executor.NewServiceScanner(s.Blacklist.Packages),
		executor.NewUserScanner(s.Baseline.Users),
		executor.NewFilePermScanner(),
		executor.NewSuidSgidScanner(),
		executor.NewWorldWriteScanner(),
		executor.NewPasswordPolicyScanner(),
		executor.NewFail2banScanner(),
		executor.NewKernelModuleScanner(),
		executor.NewMountScanner(),
		executor.NewMiscScanner(),
	}

	var allResults []executor.ScanResult

	for _, sc := range scanners {
		s.Broker.Publish(sse.Event{
			Type: "scan_progress",
			Data: `{"scanner":"` + sc.Name() + `","status":"başladı"}`,
		})

		results := sc.Scan()
		allResults = append(allResults, results...)

		// Her sonucu SSE ile yayınla
		for _, res := range results {
			data, _ := json.Marshal(res)
			s.Broker.Publish(sse.Event{Type: "scan_result", Data: string(data)})
		}

		s.Broker.Publish(sse.Event{
			Type: "scan_progress",
			Data: `{"scanner":"` + sc.Name() + `","status":"tamamlandı"}`,
		})
	}

	summary := executor.ComputeSummary(allResults)

	report := &executor.ScanReport{
		Timestamp: time.Now().Format(time.RFC3339),
		Hostname:  hostname,
		Results:   allResults,
		Summary:   summary,
	}

	s.lastResults = allResults
	s.lastReport = report

	// Tamamlandı SSE
	summaryData, _ := json.Marshal(summary)
	s.Broker.Publish(sse.Event{Type: "scan_complete", Data: string(summaryData)})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastReport == nil {
		http.Error(w, `{"error":"henüz tarama yapılmadı"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.lastReport)
}

func (s *Server) handleFix(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastResults == nil {
		http.Error(w, `{"error":"henüz tarama yapılmadı"}`, http.StatusBadRequest)
		return
	}

	result := executor.Fix(s.lastResults, id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleFixAll(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastResults == nil {
		http.Error(w, `{"error":"henüz tarama yapılmadı"}`, http.StatusBadRequest)
		return
	}

	results := executor.FixAll(s.lastResults)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleBackups(w http.ResponseWriter, r *http.Request) {
	entries, err := backup.ListBackups()
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (s *Server) handleRestore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	dir := "/var/lib/kalkani/backups/" + id

	if err := backup.Restore(dir); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true,"message":"geri yükleme tamamlandı"}`))
}
