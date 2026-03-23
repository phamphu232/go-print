package main

import (
	"github.com/getlantern/systray"
)

func main() {
	loadSetting()
	initLogger()

	handleStartup()
	handleUpdate(setting.AutoUpdate)

	systray.Run(onReady, onExit)
}
