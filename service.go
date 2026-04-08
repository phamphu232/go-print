package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/kardianos/service"
)

type program struct{}

func initService() service.Service {
	exePath, _ := os.Executable()

	svcConfig := &service.Config{
		Name:        "GoPrint",
		DisplayName: "GoPrintService",
		Description: "Go Print Service",

		WorkingDirectory: filepath.Dir(exePath),
	}

	prg := &program{}
	s, _ := service.New(prg, svcConfig)

	return s
}

// Logic for service
func (p *program) Start(s service.Service) error {
	go startServer(setting.Host, setting.Port)
	return nil
}

func (p *program) Stop(s service.Service) error {
	stopServer()
	return nil
}

func controlService(action string) {
	exePath, _ := os.Executable()

	switch runtime.GOOS {
	case "windows":
		psCommand := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs -WindowStyle Hidden", exePath, action)

		cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error: ", err)
		}

	case "darwin":
		script := fmt.Sprintf("do shell script (quoted form of \"%s\" & \" %s\") with administrator privileges", exePath, action)
		cmd := exec.Command("osascript", "-e", script)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error: ", err)
		}

	case "linux":
		cmd := exec.Command("pkexec", exePath, action)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error: ", err)
		}

	default:
		fmt.Println("Error: OS not support")
	}
}
