package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// FilePermRule, bir dosya/dizin izin kuralını tanımlar.
type FilePermRule struct {
	Path    string
	Owner   string
	Group   string
	MaxPerm os.FileMode
	Desc    string
}

// FilePermScanner, kritik dosya/dizin izinlerini kontrol eden modüldür.
type FilePermScanner struct{}

func NewFilePermScanner() *FilePermScanner { return &FilePermScanner{} }
func (s *FilePermScanner) Name() string    { return "Dosya/Dizin İzinleri" }

func (s *FilePermScanner) Scan() []ScanResult {
	rules := []FilePermRule{
		{"/etc/passwd", "root", "root", 0644, "Kullanıcı veritabanı"},
		{"/etc/shadow", "root", "shadow", 0640, "Parola hash dosyası"},
		{"/etc/group", "root", "root", 0644, "Grup veritabanı"},
		{"/etc/gshadow", "root", "shadow", 0640, "Grup parola dosyası"},
		{"/etc/hosts", "root", "root", 0644, "Host tanımları"},
		{"/etc/hostname", "root", "root", 0644, "Hostname dosyası"},
		{"/etc/fstab", "root", "root", 0644, "Disk bağlama tablosu"},
		{"/etc/issue", "root", "root", 0644, "Login banner"},
		{"/etc/issue.net", "root", "root", 0644, "Uzak login banner"},
		{"/etc/motd", "root", "root", 0644, "Günün mesajı"},
		{"/etc/securetty", "root", "root", 0600, "TTY güvenlik listesi"},
		{"/boot/grub/grub.cfg", "root", "root", 0600, "GRUB yapılandırması"},
		{"/etc/crontab", "root", "root", 0600, "Cron zamanlayıcı"},
		{"/etc/cron.hourly", "root", "root", 0700, "Saatlik cron dizini"},
		{"/etc/cron.daily", "root", "root", 0700, "Günlük cron dizini"},
		{"/etc/cron.weekly", "root", "root", 0700, "Haftalık cron dizini"},
		{"/etc/cron.monthly", "root", "root", 0700, "Aylık cron dizini"},
		{"/etc/ssh/sshd_config", "root", "root", 0600, "SSH sunucu yapılandırması"},
		{"/etc/sudoers", "root", "root", 0440, "Sudo yapılandırması"},
	}

	var results []ScanResult
	for _, rule := range rules {
		results = append(results, checkFilePerm(rule)...)
	}
	return results
}

func checkFilePerm(rule FilePermRule) []ScanResult {
	var results []ScanResult
	id := fmt.Sprintf("fileperm-%s", sanitizeID(rule.Path))

	info, err := os.Stat(rule.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		results = append(results, ScanResult{
			ID: id, Category: "Dosya İzinleri",
			Description: fmt.Sprintf("%s: %s", rule.Path, rule.Desc),
			Status: StatusWarn, CurrentValue: err.Error(),
		})
		return results
	}

	// İzin kontrolü
	perm := info.Mode().Perm()
	expectedPerm := fmt.Sprintf("%04o", rule.MaxPerm)
	actualPerm := fmt.Sprintf("%04o", perm)

	permResult := ScanResult{
		ID: id + "-perm", Category: "Dosya İzinleri",
		Description:   fmt.Sprintf("%s izni (%s)", rule.Path, rule.Desc),
		ExpectedValue: expectedPerm,
		CurrentValue:  actualPerm,
	}

	if perm&^rule.MaxPerm != 0 {
		permResult.Status = StatusFail
		permResult.FixCommand = fmt.Sprintf("chmod %s %s", expectedPerm, rule.Path)
		path := rule.Path
		maxP := rule.MaxPerm
		permResult.FixFunc = func() error {
			return os.Chmod(path, maxP)
		}
	} else {
		permResult.Status = StatusPass
	}
	results = append(results, permResult)

	// Sahiplik kontrolü — stat komutuyla (Linux)
	ownerResult := ScanResult{
		ID: id + "-owner", Category: "Dosya İzinleri",
		Description:   fmt.Sprintf("%s sahipliği (%s)", rule.Path, rule.Desc),
		ExpectedValue: rule.Owner + ":" + rule.Group,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "stat", "-c", "%U:%G", rule.Path).Output()
	if err != nil {
		ownerResult.Status = StatusWarn
		ownerResult.CurrentValue = "(kontrol edilemedi)"
	} else {
		actual := strings.TrimSpace(string(out))
		ownerResult.CurrentValue = actual
		expected := rule.Owner + ":" + rule.Group
		if actual == expected {
			ownerResult.Status = StatusPass
		} else {
			ownerResult.Status = StatusFail
			ownerResult.FixCommand = fmt.Sprintf("chown %s:%s %s", rule.Owner, rule.Group, rule.Path)
		}
	}
	results = append(results, ownerResult)

	return results
}

func sanitizeID(path string) string {
	out := ""
	for _, c := range path {
		if c == '/' || c == '.' {
			out += "_"
		} else {
			out += string(c)
		}
	}
	return out
}

