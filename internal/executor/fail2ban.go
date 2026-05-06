package executor

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Fail2banScanner, fail2ban durumunu kontrol eden modüldür.
type Fail2banScanner struct{}

func NewFail2banScanner() *Fail2banScanner { return &Fail2banScanner{} }
func (s *Fail2banScanner) Name() string    { return "Brute-Force Koruması" }

func (s *Fail2banScanner) Scan() []ScanResult {
	var results []ScanResult

	// fail2ban kurulu mu
	installed := isPackageInstalled("fail2ban")
	results = append(results, ScanResult{
		ID: "f2b-installed", Category: "Brute-Force",
		Description:   "fail2ban paketi kurulu mu",
		ExpectedValue: "kurulu",
		CurrentValue:  boolToStr(installed, "kurulu", "kurulu değil"),
		Status:        boolToStatus(installed),
		FixCommand:    "apt install -y fail2ban && systemctl enable --now fail2ban",
	})

	// fail2ban aktif mi
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", "is-active", "fail2ban").Output()
	active := err == nil && strings.TrimSpace(string(out)) == "active"

	results = append(results, ScanResult{
		ID: "f2b-active", Category: "Brute-Force",
		Description:   "fail2ban servisi aktif mi",
		ExpectedValue: "active",
		CurrentValue:  boolToStr(active, "active", "inactive"),
		Status:        boolToStatus(active),
		FixCommand:    "systemctl enable --now fail2ban",
	})

	// SSH jail kontrol
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	jailOut, err := exec.CommandContext(ctx2, "fail2ban-client", "status", "sshd").CombinedOutput()
	sshJail := err == nil && strings.Contains(string(jailOut), "sshd")

	results = append(results, ScanResult{
		ID: "f2b-ssh-jail", Category: "Brute-Force",
		Description:   "fail2ban SSH jail yapılandırması",
		ExpectedValue: "aktif",
		CurrentValue:  boolToStr(sshJail, "aktif", "yapılandırılmamış"),
		Status:        boolToStatus(sshJail),
		FixCommand:    "fail2ban jail.conf içinde [sshd] enabled = true ekle",
	})

	return results
}
