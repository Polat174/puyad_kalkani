package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// UserScanner, kullanıcı güvenlik kontrollerini yapan modüldür.
type UserScanner struct {
	Baseline map[string]string
}

func NewUserScanner(baseline map[string]string) *UserScanner {
	return &UserScanner{Baseline: baseline}
}

func (s *UserScanner) Name() string { return "Kullanıcı Güvenliği" }

func (s *UserScanner) Scan() []ScanResult {
	var results []ScanResult
	results = append(results, s.checkUID0()...)
	results = append(results, s.checkEmptyPasswords()...)
	results = append(results, s.checkUmask()...)
	return results
}

func (s *UserScanner) checkUID0() []ScanResult {
	r := ScanResult{
		ID: "user-uid0", Category: "Kullanıcı",
		Description: "Root dışında UID 0 hesabı olmamalı", ExpectedValue: "sadece root",
	}
	f, err := os.Open("/etc/passwd")
	if err != nil {
		r.Status = StatusWarn
		r.CurrentValue = err.Error()
		return []ScanResult{r}
	}
	defer f.Close()
	var uid0 []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ":")
		if len(parts) >= 3 && parts[2] == "0" && parts[0] != "root" {
			uid0 = append(uid0, parts[0])
		}
	}
	if len(uid0) > 0 {
		r.Status = StatusFail
		r.CurrentValue = strings.Join(uid0, ", ")
	} else {
		r.Status = StatusPass
		r.CurrentValue = "sadece root"
	}
	return []ScanResult{r}
}

func (s *UserScanner) checkEmptyPasswords() []ScanResult {
	r := ScanResult{
		ID: "user-empty-pw", Category: "Kullanıcı",
		Description: "Boş parolalı hesap olmamalı", ExpectedValue: "yok",
	}
	f, err := os.Open("/etc/shadow")
	if err != nil {
		r.Status = StatusWarn
		r.CurrentValue = "(shadow okunamadı — root gerekli)"
		return []ScanResult{r}
	}
	defer f.Close()
	var empty []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ":")
		if len(parts) >= 2 && parts[1] == "" {
			empty = append(empty, parts[0])
		}
	}
	if len(empty) > 0 {
		r.Status = StatusFail
		r.CurrentValue = strings.Join(empty, ", ")
		r.FixCommand = fmt.Sprintf("passwd -l %s", strings.Join(empty, " && passwd -l "))
	} else {
		r.Status = StatusPass
		r.CurrentValue = "yok"
	}
	return []ScanResult{r}
}

func (s *UserScanner) checkUmask() []ScanResult {
	expected := s.Baseline["umask"]
	if expected == "" {
		expected = "027"
	}
	r := ScanResult{
		ID: "user-umask", Category: "Kullanıcı",
		Description: "Sistem varsayılan umask değeri", ExpectedValue: expected,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "bash", "-c", "umask").CombinedOutput()
	if err != nil {
		r.Status = StatusWarn
		r.CurrentValue = "(okunamadı)"
		return []ScanResult{r}
	}
	current := strings.TrimSpace(string(out))
	current = strings.TrimLeft(current, "0")
	if current == "" {
		current = "0"
	}
	norm := strings.TrimLeft(expected, "0")
	if norm == "" {
		norm = "0"
	}
	r.CurrentValue = current
	if current == norm {
		r.Status = StatusPass
	} else {
		r.Status = StatusWarn
		r.FixCommand = fmt.Sprintf("echo 'umask %s' >> /etc/profile", expected)
	}
	return []ScanResult{r}
}
