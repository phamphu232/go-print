package main

import (
	"os"

	"github.com/getlantern/systray"
	"github.com/kardianos/service"
)

func main() {
	bootstrap()

	s := initService()

	// Handle install / uninstall
	if len(os.Args) > 1 {
		controlService(os.Args[1])
		return
	}

	if !service.Interactive() {
		s.Run()
	} else {
		systray.Run(onReady, onExit)
	}
}
