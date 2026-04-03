package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/getlantern/systray"
)

//go:embed icons/running.ico
var runningIcon []byte

//go:embed icons/stopped.ico
var stoppedIcon []byte

type AppStatus int

const (
	StatusStopped AppStatus = iota // 0
	StatusRunning                  // 1
)

var (
	menuStatus  *systray.MenuItem
	menuStart   *systray.MenuItem
	menuStop    *systray.MenuItem
	menuRestart *systray.MenuItem
	menuSetting *systray.MenuItem

	appStatus AppStatus
)

func updateUIByStatus(status AppStatus) {
	if menuStatus == nil {
		return
	}

	appStatus = status

	switch status {
	case StatusRunning:
		systray.SetIcon(runningIcon)
		systray.SetTooltip("Go Print: Running")
		menuStatus.SetTitle(fmt.Sprintf("Listen: %s:%d", setting.Host, setting.Port))

		menuStart.Hide()
		menuStop.Show()
		menuRestart.Show()
		menuSetting.Show()

	case StatusStopped:
		systray.SetIcon(stoppedIcon)
		systray.SetTooltip("Go Print: Stopped")
		menuStatus.SetTitle("Go Print: Stopped")

		menuStart.Show()
		menuStop.Hide()
		menuRestart.Hide()
		menuSetting.Hide()
	}
}

func onReady() {
	systray.SetIcon(stoppedIcon)
	systray.SetTooltip("Go Print")

	menuStatus = systray.AddMenuItem("Go Print: Stopped", "Server status")
	menuStart = systray.AddMenuItem("Start", "Start server")
	menuRestart = systray.AddMenuItem("Restart", "Restart server")
	menuStop = systray.AddMenuItem("Stop", "Stop server")
	menuSetting = systray.AddMenuItem("Setting", "Open Setting")

	mExit := systray.AddMenuItem("Exit", "Exit")

	updateUIByStatus(StatusRunning)

	// go startServer(setting.Host, setting.Port)

	go func() {
		for {
			select {
			case <-menuStatus.ClickedCh:
				if appStatus == StatusRunning {
					openBrowser(fmt.Sprintf("http://%s:%d", setting.Host, setting.Port))
				}

			case <-menuStart.ClickedCh:
				startServer(setting.Host, setting.Port)

			case <-menuStop.ClickedCh:
				stopServer()

			case <-menuRestart.ClickedCh:
				go func() {
					restartServer(setting.Host, setting.Port)
				}()

			case <-menuSetting.ClickedCh:
				openBrowser(fmt.Sprintf("http://%s:%d/setting", setting.Host, setting.Port))

			case <-mExit.ClickedCh:
				stopServer()
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}

func onExit() {}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	exec.Command(cmd, args...).Start()
}
