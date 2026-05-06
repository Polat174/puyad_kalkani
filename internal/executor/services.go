package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ServiceScanner, kara listedeki paket ve servisleri kontrol eden modüldür.
type ServiceScanner struct {
	Blacklist []string
}

// NewServiceScanner, verilen kara liste ile yeni bir tarayıcı oluşturur.
func NewServiceScanner(blacklist []string) *ServiceScanner {
	return &ServiceScanner{Blacklist: blacklist}
}

func (s *ServiceScanner) Name() string { return "Servis/Paket Kara Listesi" }

// Scan, kara listedeki paketlerin/servislerin durumunu kontrol eder.
func (s *ServiceScanner) Scan() []ScanResult {
	var results []ScanResult

	for _, pkg := range s.Blacklist {
		id := fmt.Sprintf("blacklist-%s", strings.ReplaceAll(pkg, "-", "_"))

		r := ScanResult{
			ID:            id,
			Category:      "Kara Liste",
			Description:   fmt.Sprintf("Kara liste: %s", pkg),
			ExpectedValue: "yüklü değil",
		}

		installed := isPackageInstalled(pkg)
		if installed {
			r.Status = StatusFail
			r.CurrentValue = "yüklü"
			r.FixCommand = fmt.Sprintf("apt purge -y %s || yum remove -y %s", pkg, pkg)
			pkgCopy := pkg
			r.FixFunc = func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, "apt", "purge", "-y", pkgCopy)
				if err := cmd.Run(); err != nil {
					cmd = exec.CommandContext(ctx, "yum", "remove", "-y", pkgCopy)
					return cmd.Run()
				}
				return nil
			}
		} else {
			r.Status = StatusPass
			r.CurrentValue = "yüklü değil"
		}

		// Servisin çalışıp çalışmadığını da kontrol et
		active := isServiceActive(pkg)
		if active {
			svcResult := ScanResult{
				ID:            id + "-svc",
				Category:      "Kara Liste",
				Description:   fmt.Sprintf("Kara liste servisi: %s", pkg),
				ExpectedValue: "inactive",
				CurrentValue:  "active",
				Status:        StatusFail,
				FixCommand:    fmt.Sprintf("systemctl stop %s && systemctl disable %s", pkg, pkg),
			}
			svcPkg := pkg
			svcResult.FixFunc = func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				exec.CommandContext(ctx, "systemctl", "stop", svcPkg).Run()
				return exec.CommandContext(ctx, "systemctl", "disable", svcPkg).Run()
			}
			results = append(results, svcResult)
		}

		results = append(results, r)
	}

	return results
}

// runWithTimeout, komut çalıştırır ve 3 saniye timeout uygular.
func runWithTimeout(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// isPackageInstalled, paketin yüklü olup olmadığını kontrol eder (dpkg veya rpm).
func isPackageInstalled(pkg string) bool {
	// dpkg dene
	out, err := runWithTimeout("dpkg", "-l", pkg)
	if err == nil && strings.Contains(string(out), "ii") {
		return true
	}

	// rpm dene
	_, err = runWithTimeout("rpm", "-q", pkg)
	return err == nil
}

// isServiceActive, servisin aktif olup olmadığını kontrol eder.
func isServiceActive(svc string) bool {
	out, err := runWithTimeout("systemctl", "is-active", svc)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}
