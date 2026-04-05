package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kardianos/service"
)

func statusService() service.Status {
	s := initService()
	status, _ := s.Status()
	return status
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
