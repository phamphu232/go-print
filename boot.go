package main

import (
	"log"
	"os"
	"runtime"
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

func handleStartup() {
	hostname, _ := os.Hostname()
	printers, _ := getPrinters()

	cachedPCInfo = PCInfo{
		AppVersion: "0.0.1",
		Hostname:   hostname,
		OSPlatform: runtime.GOOS,
		OSArch:     runtime.GOARCH,
		OSVersion:  getOSVersion(),
		Printers:   printers,
	}

	err := os.MkdirAll("downloads", 0755)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
	}
}
