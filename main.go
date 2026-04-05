package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/getlantern/systray"
	"github.com/gofrs/flock"
	"github.com/kardianos/service"
)

var mutex sync.Mutex

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
		lockPath := filepath.Join(os.TempDir(), "goprint_tray.lock")
		fileLock := flock.New(lockPath)
		locked, err := fileLock.TryLock()
		if err != nil || !locked {
			fmt.Println("Program is already running!")
			return
		}
		defer fileLock.Unlock()

		if !isServiceRunning() {
			controlService("install")
		}
		systray.Run(onReady, onExit)
	}
}
