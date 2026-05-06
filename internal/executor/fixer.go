package executor

import "fmt"

// FixResult, bir düzeltme işleminin sonucunu tutar.
type FixResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Fix, belirli bir tarama sonucunu düzeltir.
func Fix(results []ScanResult, id string) FixResult {
	for _, r := range results {
		if r.ID == id {
			if r.FixFunc == nil {
				return FixResult{ID: id, Success: false, Message: "bu bulgu için otomatik düzeltme mevcut değil"}
			}
			if err := r.FixFunc(); err != nil {
				return FixResult{ID: id, Success: false, Message: fmt.Sprintf("düzeltme hatası: %v", err)}
			}
			return FixResult{ID: id, Success: true, Message: "başarıyla düzeltildi"}
		}
	}
	return FixResult{ID: id, Success: false, Message: "bulgu bulunamadı"}
}

// FixAll, tüm FAIL durumundaki bulguları düzeltir.
func FixAll(results []ScanResult) []FixResult {
	var out []FixResult
	for _, r := range results {
		if r.Status == StatusFail && r.FixFunc != nil {
			out = append(out, Fix(results, r.ID))
		}
	}
	return out
}
