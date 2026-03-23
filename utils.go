package main

import (
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
)

type PrinterInfo struct {
	Name       string `json:"name"`
	ShareName  string `json:"share_name"`
	DriverName string `json:"driver_name"`
	PortName   string `json:"port_name"`
}

func getPrinters() ([]PrinterInfo, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := strings.Join([]string{
			`[Console]::OutputEncoding = [System.Text.Encoding]::UTF8`,
			`Get-Printer | Select Name, ShareName, DriverName, PortName | ConvertTo-Json -Depth 2`,
		}, ";")

		out, err := exec.Command("powershell", "-Command", cmd).Output()
		if err != nil {
			return nil, err
		}

		var printers []PrinterInfo
		data := strings.TrimSpace(string(out))

		// handle single object or array
		if strings.HasPrefix(data, "{") {
			var single PrinterInfo
			if err := json.Unmarshal([]byte(data), &single); err != nil {
				return nil, err
			}
			printers = append(printers, single)
		} else {
			if err := json.Unmarshal([]byte(data), &printers); err != nil {
				return nil, err
			}
		}

		// normalize unicode
		for i := range printers {
			printers[i].Name = normalizeText(printers[i].Name)
			printers[i].ShareName = normalizeText(printers[i].ShareName)
			printers[i].DriverName = normalizeText(printers[i].DriverName)
			printers[i].PortName = normalizeText(printers[i].PortName)
		}

		return printers, nil

	case "linux", "darwin":
		out, err := exec.Command("lpstat", "-p", "-l").Output()
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(out), "\n")
		var printers []PrinterInfo
		var current PrinterInfo

		for _, line := range lines {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "printer ") {
				if current.Name != "" {
					printers = append(printers, current)
				}

				parts := strings.Split(line, " ")
				if len(parts) >= 2 {
					current = PrinterInfo{
						Name: normalizeText(parts[1]),
					}
				}
			}

			if strings.Contains(line, "Interface:") {
				current.DriverName = normalizeText(line)
			}

			if strings.Contains(line, "device for") {
				current.PortName = normalizeText(line)
			}
		}

		// add last printer
		if current.Name != "" {
			printers = append(printers, current)
		}

		return printers, nil
	}

	return nil, nil
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
