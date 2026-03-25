package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type Setting struct {
	Host             string `json:"host"`
	Port             int    `json:"port"`
	RunAtStartup     bool   `json:"run_at_startup"`
	AutoUpdate       bool   `json:"auto_update"`
	LogRetentionDays int    `json:"log_retention_days"`
	PrintProcessor   string `json:"print_processor"`
}

var (
	settingFile = "setting.json"
	setting     Setting
)

func loadSetting() {

	if _, err := os.Stat(settingFile); os.IsNotExist(err) {
		setting = Setting{
			Host:             "127.0.0.1",
			Port:             6868,
			RunAtStartup:     true,
			AutoUpdate:       true,
			LogRetentionDays: 30,
			PrintProcessor:   "",
		}

		data, _ := json.MarshalIndent(setting, "", "    ")
		err := os.WriteFile(settingFile, data, 0644)
		if err != nil {
			log.Printf("Failed to create default setting file: %v", err)
		}
		return
	}

	file, err := os.ReadFile(settingFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &setting)
	if err != nil {
		log.Fatalf("Failed to parse setting file: %v", err)
	}
}

func saveSetting() {
	data, _ := json.MarshalIndent(setting, "", "    ")
	os.WriteFile(settingFile, data, 0644)
}

func editSetting(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	ref := r.Referer()
	processorOptions := ""
	if runtime.GOOS == "windows" {
		processorOptions = fmt.Sprintf("<option value=\"ghostscript\" %s>Ghostscript</option>", selected(setting.PrintProcessor, "ghostscript"))
		processorOptions += fmt.Sprintf("<option value=\"sumatra-pdf\" %s>Sumatra-PDF</option>", selected(setting.PrintProcessor, "sumatra-pdf"))
	} else {
		processorOptions = fmt.Sprintf("<option value=\"cups\" %s>CUPS</option>", selected(setting.PrintProcessor, "cups"))
	}

	fmt.Fprintf(
		w,
		`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Settings</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					padding: 15px;
				}
				fieldset {
					display: inline-block;
					min-width: 350px;
					max-width: 100%%;
					border: 1px solid #ccc;
					border-radius: 4px;
				}
				legend {
					font-weight: bold;
					padding: 0 10px;
				}
				table {
					width: 100%%;
				}
				input[type="text"], input[type="number"], select {
					padding: 5px;
				}
				input[type="text"] {
					width: 200px;
				}
				button {
					padding: 8px 16px;
					background-color: #007bff;
					color: white;
					border: none;
					border-radius: 4px;
					cursor: pointer;
				}
				button:hover {
					background-color: #0056b3;
				}
			</style>
		</head>
		<body>
			<fieldset>
				<legend>Settings</legend>
				<form action="/setting/update" method="post">
					<table cellpadding="5" cellspacing="0" border="0">
						<tr>
							<td>Listen host:</td>
							<td>
								<input type="text" name="host" placeholder="127.0.0.1" value="%s" title="Input: 0.0.0.0 to allow remote setting">
							</td>
						</tr>
						<tr>
							<td>Listen port:</td>
							<td><input type="number" name="port" placeholder="6868" value="%d" min="1" max="65535"></td>
						</tr>
						<tr>
							<td>Print processor:</td>
							<td><select name="print_processor">%s</select></td>
						</tr>
						<tr>
							<td>Run on startup:</td>
							<td><input type="checkbox" name="run_at_startup" %s></td>
						</tr>
						<!--
						<tr>
							<td>Auto update:</td>
							<td><input type="checkbox" name="auto_update" %s></td>
						</tr>
						-->
						<tr>
							<td>Log retention:</td>
							<td><input type="number" name="log_retention_days" placeholder="30" value="%d" min="0" max="65535" width="50px">&nbsp;days</td>
						</tr>
						<tr>
							<td></td>
							<td>
								<button type="submit">Save</button>
								%s
							</td>
						</tr>
					</table>
				</form>
			</fieldset>
		</body>
		</html>
		`,
		setting.Host,
		setting.Port,
		processorOptions,
		checked(setting.RunAtStartup),
		checked(setting.AutoUpdate),
		setting.LogRetentionDays,
		func() string {
			if status == "1" && ref != "" && (strings.Contains(ref, "localhost") || strings.Contains(ref, "127.0.0.1") || strings.Contains(ref, "0.0.0.0")) {
				return fmt.Sprintf(`<span style="color:green;">%s</span>`, "Saved successfully.")
			}
			if status == "2" && ref != "" && (strings.Contains(ref, "localhost") || strings.Contains(ref, "127.0.0.1") || strings.Contains(ref, "0.0.0.0")) {
				return fmt.Sprintf(`<span style="color:red;">%s</span>`, "Failed action.")
			}
			return ""
		}(),
	)
}

func updateSetting(w http.ResponseWriter, r *http.Request) {
	settingOld := setting

	r.ParseForm()

	newHost := r.FormValue("host")
	if newHost != "127.0.0.1" && newHost != "localhost" && newHost != "0.0.0.0" {
		newHost = "127.0.0.1"
	}

	var newPort int
	fmt.Sscan(r.FormValue("port"), &newPort)

	setting.RunAtStartup = r.FormValue("run_at_startup") == "on"
	setting.AutoUpdate = r.FormValue("auto_update") == "on"
	setting.LogRetentionDays, _ = strconv.Atoi(r.FormValue("log_retention_days"))
	setting.PrintProcessor = r.FormValue("print_processor")

	saveSetting()

	status := 1

	if newHost != settingOld.Host || newPort != settingOld.Port {
		restartServer(newHost, newPort)

		if newHost != setting.Host || newPort != setting.Port {
			status = 2
		}
	}

	newURL := fmt.Sprintf("http://%s:%d/setting/?status=%d", setting.Host, setting.Port, status)
	http.Redirect(w, r, newURL, http.StatusFound)
}

func checked(v bool) string {
	if v {
		return "checked"
	}
	return ""
}

func selected(v string, value string) string {
	if v == value {
		return "selected"
	}
	return ""
}
