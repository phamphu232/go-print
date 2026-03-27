package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func getPrinters() ([]string, error) {
	printerNames := make([]string, 0)

	switch runtime.GOOS {
	case "windows":
		cmd := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; Get-Printer | Select-Object -ExpandProperty Name | ConvertTo-Json`

		out, err := exec.Command("powershell", "-Command", cmd).Output()
		if err != nil {
			return printerNames, nil
		}

		data := strings.TrimSpace(string(out))
		if data == "" {
			return printerNames, nil
		}

		if strings.HasPrefix(data, "[") {
			var list []string
			if err := json.Unmarshal([]byte(data), &list); err == nil {
				for _, name := range list {
					printerNames = append(printerNames, normalizeText(name))
				}
			}
		} else if strings.HasPrefix(data, "\"") || !strings.Contains(data, " ") {
			var single string
			if err := json.Unmarshal([]byte(data), &single); err == nil {
				printerNames = append(printerNames, normalizeText(single))
			} else {
				printerNames = append(printerNames, normalizeText(strings.Trim(data, "\"")))
			}
		}

	case "linux", "darwin":
		out, err := exec.Command("lpstat", "-a").Output()
		if err != nil {
			return printerNames, nil
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, " ")
			if len(parts) > 0 {
				printerNames = append(printerNames, normalizeText(parts[0]))
			}
		}
	}

	return printerNames, nil
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
