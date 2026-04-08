package main

import (
	"log"
	"os"
	"path/filepath"

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

	err := runAsAdmin(exePath, action)
	if err != nil {
		log.Printf("Error: %v", err)
	}
}
