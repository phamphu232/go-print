package main

import (
	"log"
	"os"
	"runtime"
)

type PCInfo struct {
	Hostname  string   `json:"pc_id"`
	OSType    string   `json:"os_type"`
	OSVersion string   `json:"os_version"`
	Version   string   `json:"version"`
	Printers  []string `json:"printers"`
}

var cachedPCInfo PCInfo

func handleStartup() {
	hostname, _ := os.Hostname()
	printers, _ := getPrinters()

	cachedPCInfo = PCInfo{
		Hostname:  hostname,
		OSType:    runtime.GOOS,
		OSVersion: getOSVersion(),
		Version:   "0.0.1",
		Printers:  printers,
	}

	err := os.MkdirAll("downloads", 0755)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
	}
}
