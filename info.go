package main

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type PCInfo struct {
	AppVersion string   `json:"app_version"`
	Hostname   string   `json:"pc_id"`
	OSPlatform string   `json:"os_platform"`
	OSVersion  string   `json:"os_version"`
	OSArch     string   `json:"os_arch"`
	Printers   []string `json:"printers"`
}

var cachedPCInfo PCInfo

func setCachePCInfo() {
	hostname, _ := os.Hostname()
	printers, _, _ := getPrinters()

	cachedPCInfo = PCInfo{
		AppVersion: "0.0.1",
		Hostname:   hostname,
		OSPlatform: runtime.GOOS,
		OSArch:     runtime.GOARCH,
		OSVersion:  getOSVersion(),
		Printers:   printers,
	}
}

type PrinterInfo struct {
	Name          string `json:"name"`
	ShareName     string `json:"share_name"`
	PrinterState  string `json:"printer_state"`
	PrinterStatus string `json:"printer_status"`
	Status        string `json:"status"`
	StatusInfo    string `json:"status_info"`
	Default       string `json:"default"`
}

func getPCInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := map[string]any{
		"status":  1,
		"message": "OK",
		"data":    cachedPCInfo,
	}

	json.NewEncoder(w).Encode(response)
}

func getPrinterInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	_, printerInfo, _ := getPrinters()

	response := map[string]any{
		"status":  1,
		"message": "OK",
		"data":    printerInfo,
	}

	json.NewEncoder(w).Encode(response)
}

func getPrinters() ([]string, map[string]interface{}, error) {
	printerNames := make([]string, 0)
	printerDetails := make(map[string]interface{})

	switch runtime.GOOS {
	case "windows":
		psScript := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8;`
		psScript += ` Get-WmiObject -Class Win32_Printer`
		psScript += ` | Select-Object Name, ShareName, PrinterState, PrinterStatus, Status, StatusInfo, Default`
		psScript += ` | ConvertTo-Csv -NoTypeInformation`

		cmd := exec.Command("powershell", "-Command", psScript)
		out, err := cmd.Output()

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
		out, err := exec.Command("lpstat", "-p").Output()
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
				nameNormalized := normalizeText(parts[1])
				info := PrinterInfo{
					Name:          strings.TrimSpace(nameNormalized),
					ShareName:     "",
					PrinterState:  "",
					PrinterStatus: "",
					Status:        parts[3],
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
