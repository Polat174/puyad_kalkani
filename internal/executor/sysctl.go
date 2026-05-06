package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/puyad/kalkani/internal/backup"
)

// SysctlScanner, kernel sysctl parametrelerini tarayan modüldür.
type SysctlScanner struct {
	Baseline map[string]string
}

// NewSysctlScanner, verilen baseline kernel ayarlarıyla yeni bir tarayıcı oluşturur.
func NewSysctlScanner(baseline map[string]string) *SysctlScanner {
	return &SysctlScanner{Baseline: baseline}
}

func (s *SysctlScanner) Name() string { return "Kernel/Sysctl Güvenliği" }

// Scan, sysctl parametrelerini baseline ile karşılaştırır.
func (s *SysctlScanner) Scan() []ScanResult {
	var results []ScanResult

	for key, expected := range s.Baseline {
		id := fmt.Sprintf("sysctl-%s", strings.ReplaceAll(key, ".", "-"))
		actual := readSysctl(key)

		r := ScanResult{
			ID:            id,
			Category:      "Kernel",
			Description:   fmt.Sprintf("Sysctl: %s", key),
			ExpectedValue: expected,
			CurrentValue:  actual,
		}

		if actual == "" {
			r.Status = StatusWarn
			r.CurrentValue = "(okunamadı)"
		} else if strings.TrimSpace(actual) != strings.TrimSpace(expected) {
			r.Status = StatusFail
			r.FixCommand = fmt.Sprintf("sysctl -w %s=%s", key, expected)
			r.FixFunc = makeSysctlFixer(key, expected)
		} else {
			r.Status = StatusPass
		}

		results = append(results, r)
	}

	return results
}

// readSysctl, /proc/sys/ üzerinden bir sysctl değerini okur.
// Docker container'da host /proc/sys /host/proc/sys olarak bağlanır.
func readSysctl(key string) string {
	subPath := strings.ReplaceAll(key, ".", "/")
	// Önce host mount'u dene (Docker)
	for _, prefix := range []string{"/host/proc/sys/", "/proc/sys/"} {
		data, err := os.ReadFile(prefix + subPath)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}

// makeSysctlFixer, sysctl değerini düzelten bir fonksiyon oluşturur.
func makeSysctlFixer(key, value string) func() error {
	return func() error {
		// Önce sysctl.conf yedeği
		backup.Backup("/etc/sysctl.conf")

		// Anlık uygula
		cmd := exec.Command("sysctl", "-w", fmt.Sprintf("%s=%s", key, value))
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("sysctl hatası: %s — %w", string(out), err)
		}

		// Kalıcı hale getir
		confLine := fmt.Sprintf("%s = %s\n", key, value)
		content, err := os.ReadFile("/etc/sysctl.conf")
		if err != nil {
			// Dosya yoksa oluştur
			return os.WriteFile("/etc/sysctl.conf", []byte(confLine), 0644)
		}

		lines := strings.Split(string(content), "\n")
		found := false
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, key) && strings.Contains(trimmed, "=") {
				lines[i] = fmt.Sprintf("%s = %s", key, value)
				found = true
				break
			}
		}

		if !found {
			lines = append(lines, confLine)
		}

		return os.WriteFile("/etc/sysctl.conf", []byte(strings.Join(lines, "\n")), 0644)
	}
}
