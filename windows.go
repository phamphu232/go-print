//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func runAsAdmin(exePath string, args string) error {
	verbPtr, _ := windows.UTF16PtrFromString("runas")

	var targetExe string
	var targetArgs string

	switch args {
	case "run":
		targetExe = "cmd.exe"
		targetArgs = fmt.Sprintf("/s /c \" %q install && %q start \"", exePath, exePath)

	case "autostart":
		targetExe = "cmd.exe"
		targetArgs = fmt.Sprintf("/s /c \" %q stop && %q uninstall && %q install && %q start \"", exePath, exePath, exePath, exePath)

	default:
		targetExe = exePath
		targetArgs = args
	}

	exePtr, err := windows.UTF16PtrFromString(targetExe)
	if err != nil {
		return err
	}
	argsPtr, err := windows.UTF16PtrFromString(targetArgs)
	if err != nil {
		return err
	}
	cwdPtr, _ := windows.UTF16PtrFromString("")

	err = windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, windows.SW_HIDE)
	if err != nil {
		return fmt.Errorf("ShellExecute error: %w", err)
	}

	return nil
}

func hidePowerShellWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}
