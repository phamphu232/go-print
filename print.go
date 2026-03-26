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
	if fileURL, hasFileURL := requestParams["file_url"].(string); hasFileURL && fileURL != "" {
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

func buildPrintCommandWithCUPS(requestParams map[string]interface{}, filePath string) string {
	// Ref: https://www.cups.org/doc/man-lp.html
	var options []string

	printerName, _ := requestParams["printer_name"].(string)
	if printerName != "" {
		options = append(options, fmt.Sprintf("-d \"%s\"", printerName))
	}
	if copies, ok := requestParams["copies"].(int); ok && copies > 1 {
		options = append(options, fmt.Sprintf("-n %d", copies))
	}
	if pageList, ok := requestParams["page_list"].(string); ok && pageList != "" {
		options = append(options, fmt.Sprintf("-P %s", pageList))
	}
	if orientation, ok := requestParams["orientation"].(string); ok && orientation != "" {
		options = append(options, fmt.Sprintf("-o orientation-requested=%s", orientation))
	}
	if duplex, ok := requestParams["duplex"].(string); ok && duplex != "" {
		options = append(options, fmt.Sprintf("-o sides=%s", duplex))
	}
	if paperSize, ok := requestParams["paper_size"].(string); ok && paperSize != "" {
		options = append(options, fmt.Sprintf("-o media=%s", paperSize))
	}
	if otherSettings, ok := requestParams["other_settings"].(string); ok && otherSettings != "" {
		options = append(options, fmt.Sprintf("-o %s", otherSettings))
	}

	printCommand := fmt.Sprintf("lp %s \"%s\"", strings.Join(options, " "), filePath)

	return printCommand
}

func buildPrintCommandWithGhostscript(requestParams map[string]interface{}, filePath string) string {
	archNumber := strings.TrimLeft(runtime.GOARCH, "abcdefghijklmnopqrstuvwxyz")
	gsExe := "gs"
	if runtime.GOOS == "windows" {
		gsExe = fmt.Sprintf("gswin%sc.exe", archNumber)
	}

	var gsOptions []string
	gsOptions = append(gsOptions, "-dPrinted")
	gsOptions = append(gsOptions, "-dNoCancel")
	gsOptions = append(gsOptions, "-dBATCH")
	gsOptions = append(gsOptions, "-dNOPAUSE")
	gsOptions = append(gsOptions, "-dNOPROMPT")
	gsOptions = append(gsOptions, "-dNOSAFER")

	printerName, _ := requestParams["printer_name"].(string)
	if printerName != "" {
		gsOptions = append(gsOptions, "-sDEVICE=mswinpr2")
		gsOptions = append(gsOptions, fmt.Sprintf("-sOutputFile=\"%%printer%%%s\"", printerName))
	}

	if paperSize, ok := requestParams["paper_size"].(string); ok && paperSize != "" {
		gsOptions = append(gsOptions, fmt.Sprintf("-sPAPERSIZE=%s", strings.ToLower(paperSize)))
	}
	if pageList, ok := requestParams["page_list"].(string); ok && pageList != "" {
		gsOptions = append(gsOptions, pageList)
	}

	if otherOptions, ok := requestParams["other_options"].(string); ok && otherOptions != "" {
		gsOptions = append(gsOptions, otherOptions)
	}

	printCommand := fmt.Sprintf("%s %s -f \"%s\"", gsExe, strings.Join(gsOptions, " "), filePath)

	return printCommand
}

func buildPrintCommandWithSumatraPDF(requestParams map[string]interface{}, filePath string) string {
	// Ref: https://www.sumatrapdfreader.org/docs/Command-line-arguments
	var printCommand string

	var printSettings []string

	if pageList, ok := requestParams["page_list"].(string); ok && pageList != "" {
		printSettings = append(printSettings, pageList)
	}

	if orientation, ok := requestParams["orientation"].(string); ok && orientation != "" {
		printSettings = append(printSettings, orientation)
	}

	if scale, ok := requestParams["scale"].(string); ok && scale != "" {
		printSettings = append(printSettings, scale)
	}

	if color, ok := requestParams["color"].(string); ok && color != "" {
		printSettings = append(printSettings, color)
	}

	if duplex, ok := requestParams["duplex"].(string); ok && duplex != "" {
		printSettings = append(printSettings, duplex)
	}

	if bin, ok := requestParams["bin"].(string); ok && bin != "" {
		printSettings = append(printSettings, bin)
	}

	if paperSize, ok := requestParams["paper_size"].(string); ok && paperSize != "" {
		printSettings = append(printSettings, paperSize)
	}

	if copies, ok := requestParams["copies"].(int); ok && copies > 0 {
		printSettings = append(printSettings, fmt.Sprintf("%dx", copies))
	}

	if otherSettings, ok := requestParams["other_settings"].(string); ok && otherSettings != "" {
		printSettings = append(printSettings, otherSettings)
	}

	otherOptions, _ := requestParams["other_options"].(string)

	archNumber := strings.TrimLeft(runtime.GOARCH, "abcdefghijklmnopqrstuvwxyz")

	pwd, err := os.Getwd()
	exePath := fmt.Sprintf("\\bin\\SumatraPDF%s.exe", archNumber)
	if err == nil {
		exePath = filepath.Join(pwd, fmt.Sprintf("\\bin\\SumatraPDF%s.exe", archNumber))
	}

	printerName, _ := requestParams["printer_name"].(string)

	printCommand = fmt.Sprintf("%s", exePath)
	printCommand += fmt.Sprintf(" -print-to \"%s\"", printerName)
	printCommand += fmt.Sprintf(" -print-settings \"%s\"", strings.Join(printSettings, ","))
	printCommand += fmt.Sprintf(" -silent -exit-when-done %s", otherOptions)
	printCommand += fmt.Sprintf(" \"%s\"", filePath)

	return printCommand
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
