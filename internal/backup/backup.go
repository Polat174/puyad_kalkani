package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const backupRoot = "/var/lib/kalkani/backups"

// Entry, bir yedekleme girişini temsil eder.
type Entry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	FilePath  string `json:"file_path"`
	BackupDir string `json:"backup_dir"`
}

// Backup, belirtilen dosyanın yedeğini alır. Yedek dizini döner.
func Backup(filePath string) (string, error) {
	ts := time.Now().Format("20060102_150405")
	safeName := strings.ReplaceAll(filePath, "/", "_")
	dir := filepath.Join(backupRoot, ts+"__"+safeName)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("yedek dizini oluşturulamadı: %w", err)
	}

	src, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("kaynak dosya açılamadı: %w", err)
	}
	defer src.Close()

	dstPath := filepath.Join(dir, filepath.Base(filePath))
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("yedek dosya oluşturulamadı: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("dosya kopyalama hatası: %w", err)
	}

	return dir, nil
}

// Restore, yedek dizininden orijinal dosyayı geri yükler.
func Restore(backupDir string) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("yedek dizini okunamadı: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("yedek dizininde dosya bulunamadı")
	}

	// Dizin adından orijinal yolu çıkar
	base := filepath.Base(backupDir)
	parts := strings.SplitN(base, "__", 2)
	if len(parts) != 2 {
		return fmt.Errorf("geçersiz yedek dizin formatı: %s", base)
	}
	origPath := strings.ReplaceAll(parts[1], "_", "/")

	srcPath := filepath.Join(backupDir, entries[0].Name())
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("yedek dosya açılamadı: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(origPath)
	if err != nil {
		return fmt.Errorf("orijinal dosya yazılamadı: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("geri yükleme kopyalama hatası: %w", err)
	}

	return nil
}

// ListBackups, mevcut yedeklemeleri listeler.
func ListBackups() ([]Entry, error) {
	if _, err := os.Stat(backupRoot); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return nil, fmt.Errorf("yedek dizini okunamadı: %w", err)
	}

	var result []Entry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		parts := strings.SplitN(name, "__", 2)
		ts := ""
		filePath := ""
		if len(parts) == 2 {
			ts = parts[0]
			filePath = strings.ReplaceAll(parts[1], "_", "/")
		}
		result = append(result, Entry{
			ID:        name,
			Timestamp: ts,
			FilePath:  filePath,
			BackupDir: filepath.Join(backupRoot, name),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})

	return result, nil
}
