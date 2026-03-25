package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func getPCInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := map[string]any{
		"status":  1,
		"message": "OK",
		"data":    cachedPCInfo,
	}

	json.NewEncoder(w).Encode(response)
}

func Print(w http.ResponseWriter, r *http.Request) {
	requestParams := make(map[string]interface{})

	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			requestParams[key] = values[0]
		} else {
			requestParams[key] = values
		}
	}

	var jsonMap map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&jsonMap)
	if err == nil {
		for k, v := range jsonMap {
			requestParams[k] = v
		}
	}

	printCommand := ""

	status := 0
	message := ""

	callbackURL, hasRequestCallbackURL := requestParams["callback_url"].(string)
	defer func() {
		if hasRequestCallbackURL && callbackURL != "" {
			requestParams["status"] = status
			requestParams["message"] = message
			requestParams["print_command"] = printCommand

			sendCallback(callbackURL, requestParams)
		}
	}()

	jsonData, err := json.MarshalIndent(requestParams, "", "  ")
	if err != nil {
		message = fmt.Sprintf("Failed to marshal params: %v", err)
		log.Printf("%s", message)
		fmt.Fprintln(w, message)
		return
	}

	log.Printf("Print request: %s", string(jsonData))

	saveFilePath := ""
	if fileURL, ok := requestParams["file_url"].(string); ok && fileURL != "" {
		saveFilePath, err = downloadFile(fileURL, "./downloads")
		if err != nil {
			message = fmt.Sprintf("Failed to download file: %s, %v", fileURL, err)
			log.Printf("%s", message)
			fmt.Fprintln(w, message)
			return
		}
	}

	if saveFilePath == "" {
		message = "No file to print"
		log.Printf("%s", message)
		fmt.Fprintln(w, message)
		return
	}

	if printCommand, hasRequestPrintCommand := requestParams["print_command"].(string); hasRequestPrintCommand && printCommand != "" {
		printCommand = fmt.Sprintf("%s %s", printCommand, saveFilePath)
	} else {
		switch runtime.GOOS {
		case "windows":
			printProcessor, hasRequestPrintProcessor := requestParams["print_processor"].(string)
			if hasRequestPrintProcessor {
				printProcessor = strings.ToLower(printProcessor)
			}

			if hasRequestPrintProcessor && printProcessor == "sumatra-pdf" {
				printCommand = buildPrintCommandWithSumatraPDF(requestParams, saveFilePath)
			} else if hasRequestPrintProcessor && printProcessor == "ghostscript" {
				printCommand = buildPrintCommandWithGhostscript(requestParams, saveFilePath)
			} else if setting.PrintProcessor == "sumatra-pdf" {
				printCommand = buildPrintCommandWithSumatraPDF(requestParams, saveFilePath)
			} else {
				printCommand = buildPrintCommandWithGhostscript(requestParams, saveFilePath)
			}
		default:
			printCommand = buildPrintCommandWithCUPS(requestParams, saveFilePath)
		}
	}

	if printCommand != "" && saveFilePath != "" {
		out, err := exec.Command(printCommand).Output()
		if err != nil {
			message = fmt.Sprintf("Failed to print file: %s, %v", saveFilePath, err)
			log.Printf("%s", message)
			fmt.Fprintln(w, message)
			return
		}
		status = 1
		message = fmt.Sprintf("File printed successfully: %s", string(out))
		log.Printf("%s", message)
		fmt.Fprintln(w, message)
	}

	deleteAfterPrint, hasRequestDeleteAfterPrint := requestParams["delete_after_print"].(bool)
	if hasRequestDeleteAfterPrint && deleteAfterPrint {
		err := os.Remove(saveFilePath)
		if err != nil {
			message = fmt.Sprintf("Failed to delete file: %s, %v", saveFilePath, err)
			log.Printf("%s", message)
			fmt.Fprintln(w, message)
		}
	}
}

func downloadFile(url string, saveDir string) (string, error) {
	today := time.Now().Format("2006/01/02")
	todayDir := filepath.Join(saveDir, today)

	err := os.MkdirAll(todayDir, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create directory: %v", err)
	}

	fileName := filepath.Base(url)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = "downloaded_file"
	}

	destPath := filepath.Join(todayDir, fileName)

	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, ext)
		counter := 1
		for {
			newFileName := fmt.Sprintf("%s (%d)%s", baseName, counter, ext)
			destPath = filepath.Join(todayDir, newFileName)
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				break
			}
			counter++
		}
	}

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("Failed to create file: %v", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to save file: %v", err)
	}

	return destPath, nil
}

func buildPrintCommandWithCUPS(requestParams map[string]interface{}, saveFilePath string) string {
	// @Todo
	return saveFilePath
}

func buildPrintCommandWithGhostscript(requestParams map[string]interface{}, saveFilePath string) string {
	// @Todo
	return saveFilePath
}

func buildPrintCommandWithSumatraPDF(requestParams map[string]interface{}, saveFilePath string) string {
	// @Todo
	return saveFilePath
}

func sendCallback(url string, payload map[string]interface{}) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Callback] Marshal error: %v", err)
		return
	}

	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[Callback] Post error to %s: %v", url, err)
			return
		}
		defer resp.Body.Close()

		log.Printf("[Callback] Sent to %s - Status: %s", url, resp.Status)
	}()
}
