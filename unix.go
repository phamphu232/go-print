//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func runAsAdmin(exePath string, args string) error {
	switch runtime.GOOS {

	case "darwin":
		script := fmt.Sprintf("do shell script (quoted form of \"%s\" & \" %s\") with administrator privileges", exePath, args)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()

	case "linux":
		cmd := exec.Command("pkexec", exePath, args)
		return cmd.Run()

	default:
		return fmt.Errorf("Error: OS not support")
	}
}

func hidePowerShellWindow(cmd *exec.Cmd) {
	// Do nothing on non-Windows systems
}
