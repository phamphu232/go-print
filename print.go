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
	"strconv"
	"strings"
	"time"
)

func Print(w http.ResponseWriter, r *http.Request) {
	requestParams := make(map[string]interface{})

	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			requestParams[key] = values[0]
		} else {
			requestParams[key] = values
		}
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var jsonMap map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&jsonMap)
		if err == nil {
			for k, v := range jsonMap {
				requestParams[k] = v
			}
		}
	} else {
		r.ParseForm()

		for k, v := range r.Form {
			if len(v) > 0 {
				requestParams[k] = v[0]
			}
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

			jsonCallbackRequestParams, err := json.Marshal(requestParams)
			if err == nil {
				log.Printf("Callback Request params: %s", string(jsonCallbackRequestParams))
			}

			sendCallback(callbackURL, requestParams)
		}
	}()

	jsonRequestParams, err := json.Marshal(requestParams)
	if err != nil {
		message = fmt.Sprintf("Failed to marshal params: %v", err)
		log.Printf("%s", message)
		fmt.Fprintln(w, message)
		return
	}

	log.Printf("Print request: %s", string(jsonRequestParams))

	saveFilePath := ""
	if fileURL, ok := requestParams["file_url"].(string); ok && fileURL != "" {
		saveFilePath, err = downloadFile(fileURL, filepath.Join(baseDir(), "downloads"))
		if err != nil {
			message = fmt.Sprintf("⚠️ Failed to download file: %s, %v", fileURL, err)
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

	// Prefer print command from request
	pCmd, _ := requestParams["print_command"].(string)
	if pCmd != "" {
		printCommand = fmt.Sprintf("%s %s", pCmd, saveFilePath)
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

	log.Printf("Print command: %s", printCommand)
	fmt.Fprintln(w, "Print command: ", printCommand)

	if printCommand != "" {
		var cmd *exec.Cmd

		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", printCommand)
		} else {
			cmd = exec.Command("sh", "-c", printCommand)
		}

		out, err := cmd.CombinedOutput()

		if err != nil {
			message = fmt.Sprintf("⚠️ Failed to print file: %s, Error: %v", saveFilePath, err)
			log.Printf("%s", message)
			fmt.Fprintln(w, message)
			return
		}

		status = 1
		message = fmt.Sprintf("✅ File printed successfully!, 🖨️ Output: %s", string(out))
		log.Printf("%s", message)
		fmt.Fprintln(w, message)
	}

	deleteAfterStr, _ := requestParams["delete_after_print"].(string)
	deleteAfterPrint, err := strconv.ParseBool(deleteAfterStr)
	if err == nil && deleteAfterPrint {
		err := os.Remove(saveFilePath)
		if err != nil {
			message = fmt.Sprintf("Failed to delete file: %s, %v", saveFilePath, err)
			log.Printf("%s", message)
			fmt.Fprintln(w, message)
		}
	}
}

func downloadFile(url string, saveDir string) (string, error) {
	todayDir := filepath.Join(saveDir, todayDir())

	err := os.MkdirAll(todayDir, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create directory: %v", err)
	}

	fileName := filepath.Base(url)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = "unknown"
	}

	destPath := filepath.Join(todayDir, fileName)

	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, ext)
		counter := 1
		for {
			newFileName := fmt.Sprintf("%s(%d)%s", baseName, counter, ext)
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

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Get(url)

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
	if printerName == "" {
		printerName, _ = requestParams["printer_id"].(string)
	}

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
		options = append(options, fmt.Sprintf("-o %s", orientation))
	}

	if duplex, ok := requestParams["duplex"].(string); ok && duplex != "" {
		options = append(options, fmt.Sprintf("-o %s", duplex))
	}

	if paperSize, ok := requestParams["paper_size"].(string); ok && paperSize != "" {
		options = append(options, fmt.Sprintf("-o media=%s", paperSize))
	}

	if otherSettings, ok := requestParams["other_settings"].(string); ok && otherSettings != "" {
		options = append(options, otherSettings)
	}

	if otherOptions, ok := requestParams["other_options"].(string); ok && otherOptions != "" {
		options = append(options, otherOptions)
	}

	printCommand := fmt.Sprintf("lp %s \"%s\"", strings.Join(options, " "), filePath)

	return printCommand
}

func buildPrintCommandWithGhostscript(requestParams map[string]interface{}, filePath string) string {
	archNumber := strings.TrimLeft(runtime.GOARCH, "abcdefghijklmnopqrstuvwxyz")

	gsExe := filepath.Join(baseDir(), "bin", fmt.Sprintf("gswin%sc.exe", archNumber))

	var gsOptions []string
	gsOptions = append(gsOptions, "-dNOPAUSE")
	gsOptions = append(gsOptions, "-dNOPROMPT")
	gsOptions = append(gsOptions, "-dNoCancel")
	gsOptions = append(gsOptions, "-dSAFER")
	gsOptions = append(gsOptions, "-dBATCH")

	printerName, _ := requestParams["printer_name"].(string)
	if printerName == "" {
		printerName, _ = requestParams["printer_id"].(string)
	}

	if printerName != "" {
		gsOptions = append(gsOptions, "-sDEVICE=mswinpr2")
		gsOptions = append(gsOptions, fmt.Sprintf("-sOutputFile=\"%%printer%%%s\"", printerName))
	}

	if paperSize, ok := requestParams["paper_size"].(string); ok && paperSize != "" {
		gsOptions = append(gsOptions, fmt.Sprintf("-sPAPERSIZE=%s", strings.ToLower(paperSize)))
	} else if pageWith, ok := requestParams["page_width"].(string); ok && pageWith != "" {
		if pageHeight, ok := requestParams["page_height"].(string); ok && pageHeight != "" {
			gsOptions = append(gsOptions, "-dFIXEDMEDIA")
			gsOptions = append(gsOptions, "-dPDFFitPage")
			gsOptions = append(gsOptions, fmt.Sprintf("-dDEVICEWIDTHPOINTS=%s", pageWith))
			gsOptions = append(gsOptions, fmt.Sprintf("-dDEVICEHEIGHTPOINTS=%s", pageHeight))
		}
	}

	if orientation, ok := requestParams["orientation"].(string); ok && orientation != "" {
		switch strings.ToLower(orientation) {
		case "landscape":
			gsOptions = append(gsOptions, "-c \"<</Orientation 3>> setpagedevice\"")
		case "portrait":
			gsOptions = append(gsOptions, "-c \"<</Orientation 1>> setpagedevice\"")
		}
	}

	if pageList, ok := requestParams["page_list"].(string); ok && pageList != "" {
		gsOptions = append(gsOptions, fmt.Sprintf("-sPageList=%s", pageList))
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

	exePath := filepath.Join(baseDir(), "bin", fmt.Sprintf("SumatraPDF%s.exe", archNumber))

	printerName, _ := requestParams["printer_name"].(string)
	if printerName == "" {
		printerName, _ = requestParams["printer_id"].(string)
	}

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
