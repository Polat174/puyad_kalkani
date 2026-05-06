package executor

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// KernelModuleScanner, kernel modül kara listesini kontrol eden modüldür.
type KernelModuleScanner struct{}

func NewKernelModuleScanner() *KernelModuleScanner { return &KernelModuleScanner{} }
func (s *KernelModuleScanner) Name() string        { return "Kernel Modül Kara Listesi" }

// Sunumda belirtilen kara listelenecek modüller
var moduleBlacklist = []struct {
	Module string
	Desc   string
}{
	{"bluetooth", "Bluetooth protokolü"},
	{"btusb", "Bluetooth USB sürücüsü"},
	{"btrtl", "Bluetooth RTL sürücüsü"},
	{"btbcm", "Bluetooth BCM sürücüsü"},
	{"btintel", "Bluetooth Intel sürücüsü"},
	{"btmtk", "Bluetooth MTK sürücüsü"},
	{"snd", "Ses alt sistemi"},
	{"soundcore", "Ses çekirdeği"},
	{"pcspkr", "PC hoparlör"},
	{"joydev", "Joystick sürücüsü"},
	{"gameport", "Oyun portu"},
}

func (s *KernelModuleScanner) Scan() []ScanResult {
	var results []ScanResult

	// Mevcut blacklist dosyasını oku
	blacklisted := loadModprobeBlacklist()

	for _, mod := range moduleBlacklist {
		id := "kmod-" + mod.Module

		r := ScanResult{
			ID: id, Category: "Kernel Modülleri",
			Description:   "Modül kara listede: " + mod.Module + " (" + mod.Desc + ")",
			ExpectedValue: "blacklisted",
		}

		inBlacklist := blacklisted[mod.Module]
		loaded := isModuleLoaded(mod.Module)

		if inBlacklist && !loaded {
			r.Status = StatusPass
			r.CurrentValue = "kara listede & yüklü değil"
		} else if inBlacklist && loaded {
			r.Status = StatusWarn
			r.CurrentValue = "kara listede ama hâlâ yüklü"
			r.FixCommand = "rmmod " + mod.Module
		} else {
			r.Status = StatusFail
			r.CurrentValue = boolToStr(loaded, "yüklü", "yüklü değil") + " & kara listede değil"
			modName := mod.Module
			r.FixCommand = "echo 'blacklist " + modName + "' >> /etc/modprobe.d/blacklist.conf"
			r.FixFunc = func() error {
				f, err := os.OpenFile("/etc/modprobe.d/blacklist.conf", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
				_, err = f.WriteString("blacklist " + modName + "\n")
				return err
			}
		}

		results = append(results, r)
	}

	return results
}

func loadModprobeBlacklist() map[string]bool {
	result := make(map[string]bool)

	files := []string{"/etc/modprobe.d/blacklist.conf", "/etc/modprobe.d/blacklist-custom.conf"}
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(line, "blacklist ") {
				mod := strings.TrimPrefix(line, "blacklist ")
				result[strings.TrimSpace(mod)] = true
			}
		}
		f.Close()
	}

	return result
}

func isModuleLoaded(mod string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "lsmod").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == mod {
			return true
		}
	}
	return false
}
