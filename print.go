package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	jsonData, err := json.MarshalIndent(requestParams, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal params: %v", err)
		fmt.Fprintln(w, "Failed to marshal params")
		return
	}

	log.Printf("Print request: %s", string(jsonData))

	saveFilePath := ""
	if fileURL, ok := requestParams["file_url"].(string); ok && fileURL != "" {
		saveFilePath, err = downloadFile(fileURL, "./downloads")
		if err != nil {
			log.Printf("Failed to download file: %v", err)
			fmt.Fprintln(w, "Failed to download file")
			return
		}
	}

	printCmd := ""
	if saveFilePath != "" {
		if printCmd, ok := requestParams["print_cmd"].(string); ok && printCmd != "" {
			printCmd = fmt.Sprintf("%s %s", printCmd, saveFilePath)
		} else {
			printCmd = buildPrintCommand(requestParams, saveFilePath)
		}
	}

	if printCmd != "" && saveFilePath != "" {
		out, err := exec.Command(printCmd).Output()
		if err != nil {
			log.Printf("Failed to print file: %v", err)
			fmt.Fprintln(w, "Failed to print file")
			return
		}
		log.Printf("File printed successfully: %s", string(out))
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

func buildPrintCommand(requestParams map[string]interface{}, saveFilePath string) string {
	return saveFilePath
}
