package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func makeDownloadDir() {
	err := os.MkdirAll(filepath.Join(baseDir(), "downloads"), 0755)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
	}
}

func startCleanupWorker() {
	go func() {
		for {
			cleanOldFiles(filepath.Join(baseDir(), "logs"), setting.LogRetentionDays)
			cleanOldFiles(filepath.Join(baseDir(), "downloads"), setting.LogRetentionDays)

			time.Sleep(24 * time.Hour)
		}
	}()
}

func normalizeText(s string) string {
	s = strings.TrimSpace(s)

	// remove control characters
	s = strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		return r
	}, s)

	return s
}

func todayDir() string {
	return filepath.Join(time.Now().Format("2006"), time.Now().Format("01"), time.Now().Format("02"))
}

func baseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir := filepath.Dir(exePath)

	return baseDir
}

func cleanOldFiles(rootPath string, days int) {
	cutoff := time.Now().AddDate(0, 0, -days)

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if !info.IsDir() {
			if info.ModTime().Before(cutoff) {
				os.Remove(path)
			}
		}

		return nil
	})

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() && path != rootPath {
			os.Remove(path)
		}

		return nil
	})
}
