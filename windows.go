//go:build windows

package main

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func runAsAdmin(exePath string, args string) error {
	verbPtr, _ := windows.UTF16PtrFromString("runas")
	exePtr, _ := windows.UTF16PtrFromString(exePath)
	argsPtr, _ := windows.UTF16PtrFromString(args)
	cwdPtr, _ := windows.UTF16PtrFromString("")

	return windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, windows.SW_HIDE)

	// psCommand := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs -WindowStyle Hidden", exePath, action)

	// cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
	// err := cmd.Run()
}

func hidePowerShellWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}
