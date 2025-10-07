package main

import (
	"log"
	"os"
	"time"
)

var (
	Log *log.Logger
)

// InitLogger khởi tạo logger lưu log vào file logs/YYYY-MM-DD.log
func InitLogger() {
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}

	// Generate log file name based on the current date
	currentDate := time.Now().Format("2006-01-02")
	logFileName := logDir + "/" + currentDate + ".log"

	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Log = log.Default()
		return
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
}
