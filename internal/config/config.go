package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// BaselineConfig, baseline.conf dosyasından okunan güvenlik profilini tutar.
type BaselineConfig struct {
	SSH        map[string]string
	Kernel     map[string]string
	Firewall   map[string]string
	Filesystem map[string]string
	Users      map[string]string
}

// BlacklistConfig, blacklist.conf dosyasından okunan kara listeyi tutar.
type BlacklistConfig struct {
	Packages []string
}

// LoadBaseline, INI benzeri baseline.conf dosyasını parse eder.
func LoadBaseline(path string) (*BaselineConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("baseline dosyası açılamadı: %w", err)
	}
	defer f.Close()

	cfg := &BaselineConfig{
		SSH:        make(map[string]string),
		Kernel:     make(map[string]string),
		Firewall:   make(map[string]string),
		Filesystem: make(map[string]string),
		Users:      make(map[string]string),
	}

	var currentSection string
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Boş satır veya yorum
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Bölüm başlığı [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(line[1 : len(line)-1])
			continue
		}

		// Anahtar = Değer
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("satır %d: geçersiz format: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "ssh":
			cfg.SSH[key] = value
		case "kernel":
			cfg.Kernel[key] = value
		case "firewall":
			cfg.Firewall[key] = value
		case "filesystem":
			cfg.Filesystem[key] = value
		case "users":
			cfg.Users[key] = value
		default:
			return nil, fmt.Errorf("satır %d: bilinmeyen bölüm: [%s]", lineNum, currentSection)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("baseline dosyası okuma hatası: %w", err)
	}

	return cfg, nil
}

// LoadBlacklist, satır bazlı blacklist.conf dosyasını parse eder.
func LoadBlacklist(path string) (*BlacklistConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("blacklist dosyası açılamadı: %w", err)
	}
	defer f.Close()

	cfg := &BlacklistConfig{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cfg.Packages = append(cfg.Packages, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("blacklist dosyası okuma hatası: %w", err)
	}

	return cfg, nil
}
