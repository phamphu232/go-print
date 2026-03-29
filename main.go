package main

import (
	"os"

	"github.com/getlantern/systray"
	"github.com/kardianos/service"
)

type program struct{}

// Logic for service
func (p *program) Start(s service.Service) error {
	bootstrap()
	go startServer(setting.Host, setting.Port)
	return nil
}

func (p *program) Stop(s service.Service) error {
	stopServer()
	return nil
}

func main() {
	bootstrap()

	svcConfig := &service.Config{
		Name:        "GoPrintService",
		DisplayName: "Go Print Service",
	}

	prg := &program{}
	s, _ := service.New(prg, svcConfig)

	// Handle install / uninstall
	if len(os.Args) > 1 {
		service.Control(s, os.Args[1])
		return
	}

	if !service.Interactive() {
		s.Run()
	} else {
		systray.Run(onReady, onExit)
	}
}
