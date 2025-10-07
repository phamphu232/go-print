package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// StartServer khởi động HTTP server và khai báo các endpoint
func StartServer(cfg *Config) {
	mux := http.NewServeMux()
	mux.Handle("/print", http.HandlerFunc(handlePrint))
	mux.Handle("/pc-info", http.HandlerFunc(handlePCInfo))
	mux.Handle("/all-printers", http.HandlerFunc(handleAllPrinters))
	mux.Handle("/health", http.HandlerFunc(handleHealth))

	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Println("HTTP server started at", addr)
	http.ListenAndServe(addr, corsMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handlePrint(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle print request
	w.Write([]byte("Print endpoint"))
}

func handlePCInfo(w http.ResponseWriter, r *http.Request) {
	pcName, err := os.Hostname()
	if err != nil {
		http.Error(w, "Unable to retrieve PC name", http.StatusInternalServerError)
		return
	}

	osName := runtime.GOOS
	osArch := runtime.GOARCH
	osType := "Unknown"
	switch osName {
	case "windows":
		osType = "Windows"
	case "darwin":
		osType = "MacOS"
	case "linux":
		osType = "Linux"
	}

	// Get system language
	lang := "Unknown"
	if osName == "windows" {
		// Windows-specific command to get system language
		cmd := exec.Command("powershell", "-Command", "[System.Globalization.CultureInfo]::CurrentUICulture.Name")
		output, err := cmd.Output()
		if err == nil {
			lang = strings.TrimSpace(string(output))
		}
	} else {
		// Unix-based systems (Linux/MacOS)
		cmd := exec.Command("locale", "LANG")
		output, err := cmd.Output()
		if err == nil {
			lang = strings.Split(strings.TrimSpace(string(output)), ".")[0]
		}
	}

	pcInfo := map[string]string{
		"pc_name": pcName,
		"os":      osName,
		"os_type": osType,
		"arch":    osArch,
		"lang":    lang,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pcInfo)
}

func handleAllPrinters(w http.ResponseWriter, r *http.Request) {
	var printers []string
	var err error

	if runtime.GOOS == "windows" {
		// Windows: Use PowerShell to list printers
		cmd := exec.Command("powershell", "-Command", "Get-Printer | Select-Object -ExpandProperty Name")
		output, err := cmd.Output()
		if err == nil {
			printers = strings.Split(strings.TrimSpace(string(output)), "\n")
		}
	} else {
		// Unix-based systems (Linux/MacOS): Use `lpstat` to list printers
		cmd := exec.Command("lpstat", "-a")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				// Extract printer name (first word in each line)
				printer := strings.Fields(line)[0]
				printers = append(printers, printer)
			}
		}
	}

	if err != nil {
		http.Error(w, "Unable to retrieve printers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"printers": printers})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
