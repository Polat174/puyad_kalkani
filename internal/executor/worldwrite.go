package executor

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// WorldWriteScanner, yazılabilir ve sahipsiz dosyaları tarayan modüldür.
type WorldWriteScanner struct{}

func NewWorldWriteScanner() *WorldWriteScanner { return &WorldWriteScanner{} }
func (s *WorldWriteScanner) Name() string      { return "Yazılabilir/Sahipsiz Dosyalar" }

func (s *WorldWriteScanner) Scan() []ScanResult {
	var results []ScanResult

	// World-writable dosyalar
	wwFiles := findWorldWritable(15)
	wwResult := ScanResult{
		ID: "worldwrite-files", Category: "Dosya Güvenliği",
		Description:   "Herkes tarafından yazılabilir dosyalar",
		ExpectedValue: "0",
		CurrentValue:  fmt.Sprintf("%d", len(wwFiles)),
	}
	if len(wwFiles) > 0 {
		wwResult.Status = StatusWarn
		show := wwFiles
		if len(show) > 5 {
			show = show[:5]
		}
		wwResult.FixCommand = fmt.Sprintf("chmod o-w: %s ...", strings.Join(show, ", "))
	} else {
		wwResult.Status = StatusPass
	}
	results = append(results, wwResult)

	// Sahipsiz dosyalar
	orphanFiles := findOrphanFiles(15)
	orphanResult := ScanResult{
		ID: "orphan-files", Category: "Dosya Güvenliği",
		Description:   "Sahipsiz dosyalar (nouser/nogroup)",
		ExpectedValue: "0",
		CurrentValue:  fmt.Sprintf("%d", len(orphanFiles)),
	}
	if len(orphanFiles) > 0 {
		orphanResult.Status = StatusWarn
		show := orphanFiles
		if len(show) > 5 {
			show = show[:5]
		}
		orphanResult.FixCommand = fmt.Sprintf("chown root:root: %s ...", strings.Join(show, ", "))
	} else {
		orphanResult.Status = StatusPass
	}
	results = append(results, orphanResult)

	return results
}

func findWorldWritable(timeoutSec int) []string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "find", "/", "-xdev", "-user", "root", "-type", "f",
		"(", "-perm", "-0002", "-a", "!", "-perm", "-1000", ")")
	out, _ := cmd.Output()
	return parseLines(string(out))
}

func findOrphanFiles(timeoutSec int) []string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "find", "/", "-xdev", "(", "-nouser", "-o", "-nogroup", ")")
	out, _ := cmd.Output()
	return parseLines(string(out))
}

func parseLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}
