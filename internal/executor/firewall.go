package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// FirewallScanner, güvenlik duvarı durumunu kontrol eden modüldür.
type FirewallScanner struct {
	Baseline map[string]string
}

// NewFirewallScanner, verilen baseline firewall ayarlarıyla yeni bir tarayıcı oluşturur.
func NewFirewallScanner(baseline map[string]string) *FirewallScanner {
	return &FirewallScanner{Baseline: baseline}
}

func (s *FirewallScanner) Name() string { return "Güvenlik Duvarı" }

// Scan, firewall durumunu kontrol eder.
func (s *FirewallScanner) Scan() []ScanResult {
	var results []ScanResult
	results = append(results, s.checkUFWActive()...)
	results = append(results, s.checkIPTables()...)
	return results
}

func (s *FirewallScanner) checkUFWActive() []ScanResult {
	r := ScanResult{
		ID:            "fw-ufw-active",
		Category:      "Firewall",
		Description:   "UFW güvenlik duvarı aktif mi",
		ExpectedValue: "active",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ufw", "status").CombinedOutput()
	if err != nil {
		r.Status = StatusWarn
		r.CurrentValue = "(ufw bulunamadı veya zaman aşımı)"
		r.FixCommand = "apt install ufw -y && ufw enable"
		return []ScanResult{r}
	}

	output := strings.ToLower(string(out))
	if strings.Contains(output, "status: active") {
		r.Status = StatusPass
		r.CurrentValue = "active"
	} else {
		r.Status = StatusFail
		r.CurrentValue = "inactive"
		r.FixCommand = "ufw --force enable"
		r.FixFunc = func() error {
			ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel2()
			_, err := exec.CommandContext(ctx2, "ufw", "--force", "enable").CombinedOutput()
			return err
		}
	}

	return []ScanResult{r}
}

func (s *FirewallScanner) checkIPTables() []ScanResult {
	var results []ScanResult

	chains := []struct {
		chain    string
		expected string
	}{
		{"INPUT", "DROP"},
		{"FORWARD", "DROP"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "iptables", "-L", "-n").CombinedOutput()
	if err != nil {
		results = append(results, ScanResult{
			ID:           "fw-iptables-read",
			Category:     "Firewall",
			Description:  "iptables kuralları okunamadı",
			Status:       StatusWarn,
			CurrentValue: "(zaman aşımı veya hata)",
		})
		return results
	}

	output := string(out)

	for _, c := range chains {
		r := ScanResult{
			ID:            fmt.Sprintf("fw-iptables-%s", strings.ToLower(c.chain)),
			Category:      "Firewall",
			Description:   fmt.Sprintf("iptables %s varsayılan politikası", c.chain),
			ExpectedValue: c.expected,
		}

		search := fmt.Sprintf("Chain %s (policy ", c.chain)
		idx := strings.Index(output, search)
		if idx < 0 {
			r.Status = StatusWarn
			r.CurrentValue = "(bulunamadı)"
		} else {
			sub := output[idx+len(search):]
			end := strings.Index(sub, ")")
			if end > 0 {
				policy := strings.TrimSpace(sub[:end])
				r.CurrentValue = policy
				if strings.EqualFold(policy, c.expected) {
					r.Status = StatusPass
				} else {
					r.Status = StatusFail
					r.FixCommand = fmt.Sprintf("iptables -P %s %s", c.chain, c.expected)
					chainCopy := c.chain
					expectedCopy := c.expected
					r.FixFunc = func() error {
						ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel2()
						_, err := exec.CommandContext(ctx2, "iptables", "-P", chainCopy, expectedCopy).CombinedOutput()
						return err
					}
				}
			}
		}

		results = append(results, r)
	}

	return results
}
