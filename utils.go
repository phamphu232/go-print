package main

import (
	"encoding/csv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func getPrinters() ([]string, map[string]interface{}, error) {
	printerNames := make([]string, 0)
	printerDetails := make(map[string]interface{})

	switch runtime.GOOS {
	case "windows":
		psScript := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8;`
		psScript += ` Get-WmiObject -Class Win32_Printer`
		psScript += ` | Select-Object Name, ShareName, PrinterState, PrinterStatus, Status, StatusInfo, Default`
		psScript += ` | ConvertTo-Csv -NoTypeInformation`

		out, err := exec.Command("powershell", "-Command", psScript).Output()
		if err != nil {
			return printerNames, printerDetails, err
		}

		reader := csv.NewReader(strings.NewReader(string(out)))
		records, err := reader.ReadAll()
		if err != nil {
			return printerNames, printerDetails, err
		}
		if len(records) < 2 {
			return printerNames, printerDetails, nil
		}

		for _, row := range records[1:] {
			if len(row) < 7 {
				continue
			}

			info := PrinterInfo{
				Name:          strings.TrimSpace(row[0]),
				ShareName:     strings.TrimSpace(row[1]),
				PrinterState:  row[2],
				PrinterStatus: row[3],
				Status:        row[4],
				StatusInfo:    row[5],
				Default:       row[6],
			}

			nameNormalized := normalizeText(info.Name)
			printerNames = append(printerNames, nameNormalized)
			printerDetails[nameNormalized] = info
		}

	case "linux", "darwin":
		out, err := exec.Command("lpstat", "-a").Output()
		if err != nil {
			return printerNames, printerDetails, nil
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, " ")
			if len(parts) > 0 {
				nameNormalized := normalizeText(parts[0])
				info := PrinterInfo{
					Name:          strings.TrimSpace(nameNormalized),
					ShareName:     "",
					PrinterState:  "",
					PrinterStatus: "",
					Status:        "",
					StatusInfo:    "",
					Default:       "",
				}

				printerNames = append(printerNames, nameNormalized)
				printerDetails[nameNormalized] = info
			}
		}
	}

	return printerNames, printerDetails, nil
}

func getOSVersion() string {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("cmd", "/C", "ver").Output()
		if err != nil {
			return "unknown"
		}
		return strings.TrimSpace(string(out))

	case "linux":
		out, err := exec.Command("bash", "-c", "cat /etc/os-release | grep PRETTY_NAME").Output()
		if err != nil {
			return "unknown"
		}

		s := string(out)
		s = strings.Split(s, "=")[1]
		return strings.Trim(s, "\"\n")

	case "darwin":
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err != nil {
			return "unknown"
		}
		return "macOS " + strings.TrimSpace(string(out))

	default:
		return "unknown"
	}
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
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
