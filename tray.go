package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
)

//go:embed icons/running.ico
var runningIcon []byte

//go:embed icons/stopped.ico
var stoppedIcon []byte

var (
	menuStatus  *systray.MenuItem
	menuStart   *systray.MenuItem
	menuStop    *systray.MenuItem
	menuRestart *systray.MenuItem
	menuSetting *systray.MenuItem
)

var serviceStatus bool

func updateTrayStatus() {
	isRunning := isServiceRunning()

	if isRunning == serviceStatus {
		return
	}

	serviceStatus = isRunning

	switch isRunning {
	case true:
		systray.SetIcon(runningIcon)
		systray.SetTooltip("Go Print: Running")
		menuStatus.SetTitle(fmt.Sprintf("Listen: %s:%d", setting.Host, setting.Port))

		menuStart.Hide()
		menuStop.Show()
		menuRestart.Show()
		menuSetting.Show()

	case false:
		systray.SetIcon(stoppedIcon)
		systray.SetTooltip("Go Print: Stopped")
		menuStatus.SetTitle("Go Print: Stopped")

		menuStart.Show()
		menuStop.Hide()
		menuRestart.Hide()
		menuSetting.Hide()
	}
}

func isServiceRunning() bool {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/ping", setting.Host, setting.Port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
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

	go func() {
		for {
			time.Sleep(2 * time.Second)
			updateTrayStatus()
		}
	}()

	go func() {
		for {
			select {
			case <-menuStatus.ClickedCh:
				if serviceStatus {
					openBrowser(fmt.Sprintf("http://%s:%d", setting.Host, setting.Port))
				}

			case <-menuStart.ClickedCh:
				if !isServiceRunning() {
					controlService("start")
				}

			case <-menuStop.ClickedCh:
				if isServiceRunning() {
					controlService("stop")
				}

			case <-menuRestart.ClickedCh:
				controlService("restart")

			case <-menuSetting.ClickedCh:
				openBrowser(fmt.Sprintf("http://%s:%d/setting", setting.Host, setting.Port))

			case <-mExit.ClickedCh:
				if isServiceRunning() {
					controlService("stop")
				}
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
