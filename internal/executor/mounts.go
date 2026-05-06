package executor

import (
	"bufio"
	"os"
	"strings"
)

// MountScanner, disk bağlama seçeneklerini kontrol eden modüldür.
type MountScanner struct{}

func NewMountScanner() *MountScanner { return &MountScanner{} }
func (s *MountScanner) Name() string { return "Disk Bağlama Kontrolü" }

func (s *MountScanner) Scan() []ScanResult {
	var results []ScanResult

	mounts := parseMounts()

	// /tmp kontrolleri
	results = append(results, checkMountOption(mounts, "/tmp", "noexec")...)
	results = append(results, checkMountOption(mounts, "/tmp", "nosuid")...)
	results = append(results, checkMountOption(mounts, "/tmp", "nodev")...)

	// /dev/shm kontrolleri
	results = append(results, checkMountOption(mounts, "/dev/shm", "noexec")...)
	results = append(results, checkMountOption(mounts, "/dev/shm", "nosuid")...)
	results = append(results, checkMountOption(mounts, "/dev/shm", "nodev")...)

	return results
}

type mountEntry struct {
	Device  string
	Point   string
	FSType  string
	Options []string
}

func parseMounts() []mountEntry {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil
	}
	defer f.Close()

	var entries []mountEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) >= 4 {
			entries = append(entries, mountEntry{
				Device:  fields[0],
				Point:   fields[1],
				FSType:  fields[2],
				Options: strings.Split(fields[3], ","),
			})
		}
	}
	return entries
}

func checkMountOption(mounts []mountEntry, point, option string) []ScanResult {
	id := "mount-" + sanitizeID(point) + "-" + option

	r := ScanResult{
		ID: id, Category: "Disk Bağlama",
		Description:   point + " — " + option + " seçeneği",
		ExpectedValue: option,
	}

	// Mount noktasını bul
	var found *mountEntry
	for i := range mounts {
		if mounts[i].Point == point {
			found = &mounts[i]
			break
		}
	}

	if found == nil {
		r.Status = StatusWarn
		r.CurrentValue = point + " ayrı bölüm olarak bağlı değil"
		r.FixCommand = "fstab'a ayrı " + point + " bölümü ekleyin"
		return []ScanResult{r}
	}

	hasOption := false
	for _, opt := range found.Options {
		if opt == option {
			hasOption = true
			break
		}
	}

	if hasOption {
		r.Status = StatusPass
		r.CurrentValue = option + " aktif"
	} else {
		r.Status = StatusFail
		r.CurrentValue = option + " eksik"
		r.FixCommand = "mount -o remount," + option + " " + point
	}

	return []ScanResult{r}
}
