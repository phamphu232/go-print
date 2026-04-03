package main

import "github.com/kardianos/service"

type program struct{}

func initService() service.Service {
	svcConfig := &service.Config{
		Name:        "GoPrint",
		DisplayName: "GoPrintService",
		Description: "Go Print Service",
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

func statusService() service.Status {
	s := initService()
	status, _ := s.Status()
	return status
}

func controlService(action string) {
	s := initService()
	service.Control(s, action)
}
