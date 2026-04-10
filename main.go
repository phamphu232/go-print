package main

import (
	"os"
	"path/filepath"

	"github.com/getlantern/systray"
	"github.com/gofrs/flock"
	"github.com/kardianos/service"
)

func main() {
	bootstrap()

	s := initService()

	// Handle install / uninstall
	if len(os.Args) > 1 {
		service.Control(s, os.Args[1])
		return
	}

	if !service.Interactive() {
		s.Run()
	} else {
		exePath, _ := os.Executable()
		lockPath := filepath.Join(filepath.Dir(exePath), ".tray.lock")
		fileLock := flock.New(lockPath)
		locked, err := fileLock.TryLock()
		if err != nil || !locked {
			return
		}
		defer fileLock.Unlock()

		if isServiceRunning() && !setting.RunAtStartup {
			controlService("stop")
		} else if !isServiceRunning() && setting.RunAtStartup {
			controlService("run")
		}

		systray.Run(onReady, onExit)
	}
}
