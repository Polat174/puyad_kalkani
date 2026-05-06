package executor

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// MiscScanner, çeşitli güvenlik kontrollerini yapan modüldür (DEP, shadow, GRUB).
type MiscScanner struct{}

func NewMiscScanner() *MiscScanner { return &MiscScanner{} }
func (s *MiscScanner) Name() string { return "Çeşitli Güvenlik Kontrolleri" }

func (s *MiscScanner) Scan() []ScanResult {
	var results []ScanResult
	results = append(results, s.checkDEP()...)
	results = append(results, s.checkShadowed()...)
	results = append(results, s.checkGRUBPassword()...)
	return results
}

// checkDEP, NX/DEP korumasını kontrol eder.
func (s *MiscScanner) checkDEP() []ScanResult {
	r := ScanResult{
		ID: "dep-nx", Category: "Bellek Güvenliği",
		Description:   "DEP/NX (Data Execution Prevention) aktif mi",
		ExpectedValue: "active",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "dmesg").Output()

	if err != nil {
		// dmesg okunamazsa /proc/cpuinfo'dan kontrol et
		cpuInfo, err2 := os.ReadFile("/proc/cpuinfo")
		if err2 == nil && strings.Contains(string(cpuInfo), " nx ") {
			r.Status = StatusPass
			r.CurrentValue = "active (cpuinfo nx flag)"
			return []ScanResult{r}
		}
		r.Status = StatusWarn
		r.CurrentValue = "(kontrol edilemedi)"
		return []ScanResult{r}
	}

	if strings.Contains(string(out), "NX (Execute Disable) protection: active") {
		r.Status = StatusPass
		r.CurrentValue = "active"
	} else if strings.Contains(string(out), "nx") || strings.Contains(string(out), "NX") {
		r.Status = StatusPass
		r.CurrentValue = "active (NX tespit edildi)"
	} else {
		r.Status = StatusWarn
		r.CurrentValue = "tespit edilemedi"
		r.FixCommand = "GRUB'a noexec=on ekle: GRUB_CMDLINE_LINUX_DEFAULT"
	}

	return []ScanResult{r}
}

// checkShadowed, /etc/passwd'da hash bulunan kullanıcıları kontrol eder.
func (s *MiscScanner) checkShadowed() []ScanResult {
	r := ScanResult{
		ID: "shadow-sync", Category: "Kullanıcı",
		Description:   "Tüm kullanıcılar shadowed olmalı",
		ExpectedValue: "tümü shadowed",
	}

	f, err := os.Open("/etc/passwd")
	if err != nil {
		r.Status = StatusWarn
		r.CurrentValue = err.Error()
		return []ScanResult{r}
	}
	defer f.Close()

	var unshadowed []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ":")
		if len(parts) >= 2 && parts[1] != "x" && parts[1] != "*" && parts[1] != "!" {
			unshadowed = append(unshadowed, parts[0])
		}
	}

	if len(unshadowed) > 0 {
		r.Status = StatusFail
		r.CurrentValue = strings.Join(unshadowed, ", ")
		r.FixCommand = "pwconv"
		r.FixFunc = func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return exec.CommandContext(ctx, "pwconv").Run()
		}
	} else {
		r.Status = StatusPass
		r.CurrentValue = "tümü shadowed"
	}

	return []ScanResult{r}
}

// checkGRUBPassword, GRUB parola korumasını kontrol eder.
func (s *MiscScanner) checkGRUBPassword() []ScanResult {
	r := ScanResult{
		ID: "grub-password", Category: "Önyükleyici",
		Description:   "GRUB parola koruması",
		ExpectedValue: "yapılandırılmış",
	}

	// /etc/grub.d/40_custom dosyasında superusers ve password_pbkdf2 kontrolü
	content, err := os.ReadFile("/etc/grub.d/40_custom")
	if err != nil {
		r.Status = StatusWarn
		r.CurrentValue = "(dosya okunamadı)"
		r.FixCommand = "grub-mkpasswd-pbkdf2 ile parola oluşturup /etc/grub.d/40_custom'a ekle"
		return []ScanResult{r}
	}

	text := string(content)
	hasSuperusers := strings.Contains(text, "superusers")
	hasPassword := strings.Contains(text, "password_pbkdf2")

	if hasSuperusers && hasPassword {
		r.Status = StatusPass
		r.CurrentValue = "yapılandırılmış"
	} else {
		r.Status = StatusFail
		r.CurrentValue = "yapılandırılmamış"
		r.FixCommand = "grub-mkpasswd-pbkdf2 ile parola oluştur, /etc/grub.d/40_custom'a ekle, update-grub çalıştır"
	}

	return []ScanResult{r}
}
