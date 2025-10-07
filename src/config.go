package main

import (
	"gopkg.in/ini.v1"
	"os"
)

// Config struct for application config
type Config struct {
	Port        int
	StartOnBoot bool
	AutoUpdate  bool
}

// LoadConfig đọc config từ file config.ini, nếu chưa có sẽ tạo file mặc định
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:        6868,
		StartOnBoot: true,
		AutoUpdate:  true,
	}
	if _, err := os.Stat("config.ini"); os.IsNotExist(err) {
		// Create default config.ini if not exist
		file, _ := os.Create("config.ini")
		file.WriteString("[General]\nport = 6868\nstart_on_boot = true\nauto_update = true\n")
		file.Close()
	}
	inicfg, err := ini.Load("config.ini")
	if err != nil {
		return cfg, err
	}
	section := inicfg.Section("General")
	cfg.Port = section.Key("port").MustInt(6868)
	cfg.StartOnBoot = section.Key("start_on_boot").MustBool(true)
	cfg.AutoUpdate = section.Key("auto_update").MustBool(true)
	return cfg, nil
}