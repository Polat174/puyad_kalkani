package executor

import (
	"os"
	"strings"
)

// PasswordPolicyScanner, parola politikası yapılandırmasını kontrol eden modüldür.
type PasswordPolicyScanner struct{}

func NewPasswordPolicyScanner() *PasswordPolicyScanner { return &PasswordPolicyScanner{} }
func (s *PasswordPolicyScanner) Name() string          { return "Parola Politikası" }

func (s *PasswordPolicyScanner) Scan() []ScanResult {
	var results []ScanResult

	// libpam-pwquality kurulu mu
	pwqInstalled := isPackageInstalled("libpam-pwquality")
	results = append(results, ScanResult{
		ID: "pw-pwquality", Category: "Parola Politikası",
		Description:   "libpam-pwquality paketi kurulu mu",
		ExpectedValue: "kurulu",
		CurrentValue:  boolToStr(pwqInstalled, "kurulu", "kurulu değil"),
		Status:        boolToStatus(pwqInstalled),
		FixCommand:    "apt install -y libpam-pwquality",
	})

	// /etc/pam.d/common-password kontrolü
	content, err := os.ReadFile("/etc/pam.d/common-password")
	if err != nil {
		results = append(results, ScanResult{
			ID: "pw-pam-read", Category: "Parola Politikası",
			Description:  "PAM common-password dosyası okunamadı",
			Status:       StatusWarn,
			CurrentValue: err.Error(),
		})
		return results
	}

	pamContent := string(content)

	// minlen kontrolü
	results = append(results, checkPAMParam(pamContent, "minlen", "8"))
	// dcredit kontrolü
	results = append(results, checkPAMParam(pamContent, "dcredit", "-1"))
	// ucredit kontrolü
	results = append(results, checkPAMParam(pamContent, "ucredit", "-1"))
	// lcredit kontrolü
	results = append(results, checkPAMParam(pamContent, "lcredit", "-1"))
	// ocredit kontrolü
	results = append(results, checkPAMParam(pamContent, "ocredit", "-1"))
	// retry kontrolü
	results = append(results, checkPAMParam(pamContent, "retry", "3"))

	// remember kontrolü (parola tekrar kullanımı)
	results = append(results, checkPAMParam(pamContent, "remember", "2"))

	return results
}

func checkPAMParam(content, param, expected string) ScanResult {
	r := ScanResult{
		ID: "pw-" + param, Category: "Parola Politikası",
		Description:   "PAM parola politikası: " + param,
		ExpectedValue: expected,
	}

	// pam_pwquality.so veya pam_unix.so satırında parametreyi ara
	found := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, param+"=") || strings.Contains(line, param+" =") {
			// Değeri çıkar
			idx := strings.Index(line, param+"=")
			if idx < 0 {
				idx = strings.Index(line, param+" =")
			}
			if idx >= 0 {
				sub := line[idx+len(param):]
				sub = strings.TrimLeft(sub, "= ")
				val := strings.Fields(sub)
				if len(val) > 0 {
					r.CurrentValue = val[0]
					if val[0] == expected {
						r.Status = StatusPass
					} else {
						r.Status = StatusFail
					}
					found = true
					break
				}
			}
		}
	}

	if !found {
		r.Status = StatusFail
		r.CurrentValue = "(tanımsız)"
		r.FixCommand = "pam_pwquality.so satırına " + param + "=" + expected + " ekle"
	}

	return r
}

func boolToStr(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

func boolToStatus(b bool) Status {
	if b {
		return StatusPass
	}
	return StatusFail
}
