package sshguard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/puyad/kalkani/internal/backup"
	"github.com/puyad/kalkani/internal/executor"
)

const sshdConfigPath = "/etc/ssh/sshd_config"

// SSHScanner, SSH sunucu yapılandırmasını tarayan modüldür.
type SSHScanner struct {
	Baseline map[string]string
}

// NewSSHScanner, verilen baseline SSH ayarlarıyla yeni bir tarayıcı oluşturur.
func NewSSHScanner(baseline map[string]string) *SSHScanner {
	return &SSHScanner{Baseline: baseline}
}

func (s *SSHScanner) Name() string { return "SSH Güvenliği" }

// Scan, sshd_config dosyasını baseline ile karşılaştırır.
func (s *SSHScanner) Scan() []executor.ScanResult {
	current, err := parseSSHDConfig(sshdConfigPath)
	if err != nil {
		return []executor.ScanResult{{
			ID:          "ssh-config-read",
			Category:    "SSH",
			Description: "sshd_config dosyası okunamadı",
			Status:      executor.StatusWarn,
			CurrentValue: err.Error(),
		}}
	}

	var results []executor.ScanResult

	for key, expected := range s.Baseline {
		actual, found := current[key]
		id := fmt.Sprintf("ssh-%s", strings.ToLower(key))

		r := executor.ScanResult{
			ID:            id,
			Category:      "SSH",
			Description:   fmt.Sprintf("SSH: %s", key),
			ExpectedValue: expected,
		}

		if !found {
			r.Status = executor.StatusWarn
			r.CurrentValue = "(tanımsız)"
			r.FixCommand = fmt.Sprintf("echo '%s %s' >> %s && systemctl restart sshd", key, expected, sshdConfigPath)
			r.FixFunc = makeSSHFixer(key, expected, false)
		} else if !strings.EqualFold(actual, expected) {
			r.Status = executor.StatusFail
			r.CurrentValue = actual
			r.FixCommand = fmt.Sprintf("sed -i 's/^%s .*/%s %s/' %s && systemctl restart sshd", key, key, expected, sshdConfigPath)
			r.FixFunc = makeSSHFixer(key, expected, true)
		} else {
			r.Status = executor.StatusPass
			r.CurrentValue = actual
		}

		results = append(results, r)
	}

	return results
}

// parseSSHDConfig, sshd_config dosyasını key-value olarak parse eder.
func parseSSHDConfig(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result, scanner.Err()
}

// makeSSHFixer, SSH ayarını düzelten bir fonksiyon oluşturur.
func makeSSHFixer(key, value string, exists bool) func() error {
	return func() error {
		// Önce yedek al
		if _, err := backup.Backup(sshdConfigPath); err != nil {
			return fmt.Errorf("yedek alınamadı: %w", err)
		}

		content, err := os.ReadFile(sshdConfigPath)
		if err != nil {
			return err
		}

		lines := strings.Split(string(content), "\n")
		newLine := fmt.Sprintf("%s %s", key, value)
		found := false

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Aktif veya yorum satırı olarak mevcut
			if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, "#"+key+" ") || strings.HasPrefix(trimmed, "# "+key+" ") {
				lines[i] = newLine
				found = true
				break
			}
		}

		if !found {
			lines = append(lines, newLine)
		}

		return os.WriteFile(sshdConfigPath, []byte(strings.Join(lines, "\n")), 0644)
	}
}
