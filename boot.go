package main

import (
	"os"
	"runtime"
)

type PCInfo struct {
	Hostname  string        `json:"pc_id"`
	OSType    string        `json:"os_type"`
	OSVersion string        `json:"os_version"`
	Printers  []PrinterInfo `json:"printers"`
}

var cachedPCInfo PCInfo

func handleStartup() {
	hostname, _ := os.Hostname()
	printers, _ := getPrinters()

	cachedPCInfo = PCInfo{
		Hostname:  hostname,
		OSType:    runtime.GOOS,
		OSVersion: getOSVersion(),
		Printers:  printers,
	}
}
