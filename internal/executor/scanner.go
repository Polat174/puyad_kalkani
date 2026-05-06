package executor

import "encoding/json"

// Status, bir tarama sonucunun durumunu belirtir.
type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
	StatusWarn Status = "WARN"
	StatusInfo Status = "INFO"
)

// ScanResult, tek bir güvenlik kontrolünün sonucunu tutar.
type ScanResult struct {
	ID            string `json:"id"`
	Category      string `json:"category"`
	Description   string `json:"description"`
	Status        Status `json:"status"`
	CurrentValue  string `json:"current_value"`
	ExpectedValue string `json:"expected_value"`
	FixCommand    string `json:"fix_command,omitempty"`
	FixFunc       func() error `json:"-"`
}

// ScanReport, tüm tarama sonuçlarını içerir.
type ScanReport struct {
	Timestamp string       `json:"timestamp"`
	Hostname  string       `json:"hostname"`
	Results   []ScanResult `json:"results"`
	Summary   Summary      `json:"summary"`
}

// Summary, tarama istatistiklerini tutar.
type Summary struct {
	Total int `json:"total"`
	Pass  int `json:"pass"`
	Fail  int `json:"fail"`
	Warn  int `json:"warn"`
	Info  int `json:"info"`
	Score int `json:"score"` // 0-100 arası güvenlik puanı
}

// Scanner, bir güvenlik tarayıcı arayüzü.
type Scanner interface {
	Name() string
	Scan() []ScanResult
}

// ComputeSummary, tarama sonuçlarından özet istatistikleri hesaplar.
func ComputeSummary(results []ScanResult) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		switch r.Status {
		case StatusPass:
			s.Pass++
		case StatusFail:
			s.Fail++
		case StatusWarn:
			s.Warn++
		case StatusInfo:
			s.Info++
		}
	}
	scoreable := s.Pass + s.Fail
	if scoreable > 0 {
		s.Score = (s.Pass * 100) / scoreable
	}
	return s
}

// ToJSON, ScanReport'u JSON formatına çevirir.
func (r *ScanReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
