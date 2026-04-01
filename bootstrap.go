package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
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

func bootstrap() {
	loadSetting()
	initLogger()
	cachePCInfo()
	makeDownloadDir()
	startCleanupWorker()
}

func cachePCInfo() {
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

func makeDownloadDir() {
	err := os.MkdirAll(filepath.Join(baseDir(), "downloads"), 0755)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
	}
}

func startCleanupWorker() {
	go func() {
		for {
			cleanOldFiles(filepath.Join(baseDir(), "logs"), setting.LogRetentionDays)
			cleanOldFiles(filepath.Join(baseDir(), "downloads"), setting.LogRetentionDays)

			time.Sleep(24 * time.Hour)
		}
	}()
}
