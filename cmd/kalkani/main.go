package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/puyad/kalkani/internal/api"
	"github.com/puyad/kalkani/internal/config"
)

func main() {
	// Config dosyalarının yolunu belirle
	exe, _ := os.Executable()
	baseDir := filepath.Dir(exe)
	configDir := filepath.Join(baseDir, "configs")

	// Eğer configs/ bulunamazsa, çalışma dizinine bak
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		configDir = "configs"
	}

	baselinePath := filepath.Join(configDir, "baseline.conf")
	blacklistPath := filepath.Join(configDir, "blacklist.conf")

	// Config yükle
	baseline, err := config.LoadBaseline(baselinePath)
	if err != nil {
		log.Fatalf("❌ Baseline yüklenemedi: %v", err)
	}
	log.Printf("✅ Baseline yüklendi: SSH=%d, Kernel=%d, Firewall=%d, Users=%d kural",
		len(baseline.SSH), len(baseline.Kernel), len(baseline.Firewall), len(baseline.Users))

	blacklist, err := config.LoadBlacklist(blacklistPath)
	if err != nil {
		log.Fatalf("❌ Blacklist yüklenemedi: %v", err)
	}
	log.Printf("✅ Blacklist yüklendi: %d paket/servis", len(blacklist.Packages))

	// API sunucusu
	srv := api.NewServer(baseline, blacklist)
	router := api.NewRouter(srv)

	// Frontend statik dosyalarını serve et
	frontendDir := filepath.Join(baseDir, "frontend", "dist")
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		frontendDir = filepath.Join("frontend", "dist")
	}
	if _, err := os.Stat(frontendDir); err == nil {
		fs := http.FileServer(http.Dir(frontendDir))
		router.Handle("/*", fs)
		log.Printf("✅ Frontend: %s", frontendDir)
	} else {
		log.Printf("⚠️  Frontend dist bulunamadı: %s (sadece API çalışacak)", frontendDir)
	}

	// HTTP sunucusu
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		fmt.Println()
		log.Printf("🛡️  PUYAD Kalkanı — Linux Hardening Sistemi")
		log.Printf("🌐 Sunucu başlatılıyor: http://localhost%s", addr)
		log.Printf("📡 API: http://localhost%s/api/scan", addr)
		log.Println("⏳ Çıkmak için Ctrl+C")
		fmt.Println()

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Sunucu hatası: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Sunucu kapatılıyor...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Kapatma hatası: %v", err)
	}
	log.Println("✅ Sunucu düzgün şekilde kapatıldı")
}
