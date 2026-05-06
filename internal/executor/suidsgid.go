package executor

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SuidSgidScanner, SUID/SGID dosyalarını tarayan modüldür.
type SuidSgidScanner struct{}

func NewSuidSgidScanner() *SuidSgidScanner { return &SuidSgidScanner{} }
func (s *SuidSgidScanner) Name() string    { return "SUID/SGID Kontrolü" }

// knownSafeSUID, sistemde olması beklenen güvenli SUID dosyaları.
var knownSafeSUID = map[string]bool{
	"/usr/bin/passwd":       true,
	"/usr/bin/chfn":         true,
	"/usr/bin/chsh":         true,
	"/usr/bin/gpasswd":      true,
	"/usr/bin/newgrp":       true,
	"/usr/bin/sudo":         true,
	"/usr/bin/su":           true,
	"/usr/bin/mount":        true,
	"/usr/bin/umount":       true,
	"/usr/lib/dbus-1.0/dbus-daemon-launch-helper": true,
	"/usr/lib/openssh/ssh-keysign":                true,
	"/usr/bin/crontab":      true,
	"/usr/bin/pkexec":       true,
	"/usr/bin/fusermount3":  true,
	"/usr/bin/fusermount":   true,
	"/usr/sbin/unix_chkpwd": true,
	"/usr/sbin/pam_timestamp_check": true,
}

func (s *SuidSgidScanner) Scan() []ScanResult {
	var results []ScanResult

	// SUID tarama
	suidFiles := findPermFiles("-4000", 15)
	unknownSUID := 0
	var unknownList []string

	for _, f := range suidFiles {
		if !knownSafeSUID[f] {
			unknownSUID++
			if len(unknownList) < 10 {
				unknownList = append(unknownList, f)
			}
		}
	}

	suidResult := ScanResult{
		ID: "suid-unknown", Category: "SUID/SGID",
		Description:   "Bilinmeyen SUID dosyaları",
		ExpectedValue: "0",
		CurrentValue:  fmt.Sprintf("%d", unknownSUID),
	}

	if unknownSUID > 0 {
		suidResult.Status = StatusWarn
		if len(unknownList) > 0 {
			suidResult.FixCommand = "Kontrol edin: " + strings.Join(unknownList, ", ")
		}
	} else {
		suidResult.Status = StatusPass
	}
	results = append(results, suidResult)

	// SGID tarama
	sgidFiles := findPermFiles("-2000", 15)
	sgidResult := ScanResult{
		ID: "sgid-count", Category: "SUID/SGID",
		Description:   "SGID dosya sayısı",
		ExpectedValue: "bilgi",
		CurrentValue:  fmt.Sprintf("%d dosya", len(sgidFiles)),
		Status:        StatusInfo,
	}
	results = append(results, sgidResult)

	return results
}

func findPermFiles(perm string, timeoutSec int) []string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "find", "/", "-xdev", "-perm", perm, "-type", "f")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var files []string
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		f := strings.TrimSpace(sc.Text())
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}
