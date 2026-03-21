package main

import (
	"github.com/getlantern/systray"
)

func main() {
	loadSetting()
	initLogger()

	handleStartup(setting.RunAtStartup)
	handleUpdate(setting.AutoUpdate)

	systray.Run(onReady, onExit)
}
