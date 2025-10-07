package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/getlantern/systray"
)

// StartTray khởi tạo tray icon và các menu
func StartTray(cfg *Config) {
	go func() {
		// Start the tray menu
		systray.Run(onReady(cfg), onExit)
	}()
}

func onReady(cfg *Config) func() {
	return func() {
		// Set tray icon and title
		if len(getIcon()) == 0 {
			fmt.Println("Icon data is empty. Tray icon will not be displayed.")
		} else {
			fmt.Println("Icon data loaded successfully.")
		}
		systray.SetTitle("Print Adapter")
		systray.SetTooltip("Print Adapter is running")

		// Add menu items
		mRestart := systray.AddMenuItem("Restart Service", "Restart the service")
		mStop := systray.AddMenuItem("Stop Service", "Stop the service")
		mConfig := systray.AddMenuItem("Config", "Open config file")
		mLogs := systray.AddMenuItem("Open Logs Folder", "Open logs folder")
		mStartOnBoot := systray.AddMenuItemCheckbox("Start on boot", "Enable/Disable start on boot", cfg.StartOnBoot)
		mAutoUpdate := systray.AddMenuItemCheckbox("Auto update", "Enable/Disable auto update", cfg.AutoUpdate)
		mAbout := systray.AddMenuItem("About", "About this application")
		mQuit := systray.AddMenuItem("Exit", "Quit the application")

		// Handle menu item clicks
		go func() {
			for {
				select {
				case <-mRestart.ClickedCh:
					fmt.Println("Restarting service...")
					// Add logic to restart the service here
				case <-mStop.ClickedCh:
					fmt.Println("Stopping service...")
					// Add logic to stop the service here
				case <-mConfig.ClickedCh:
					openFile("config.ini") // Open the config file
				case <-mLogs.ClickedCh:
					openFolder("logs") // Open the logs folder
				case <-mStartOnBoot.ClickedCh:
					cfg.StartOnBoot = !cfg.StartOnBoot
					mStartOnBoot.Check()
					fmt.Println("Start on boot:", cfg.StartOnBoot)
					// Save the updated config if needed
				case <-mAutoUpdate.ClickedCh:
					cfg.AutoUpdate = !cfg.AutoUpdate
					mAutoUpdate.Check()
					fmt.Println("Auto update:", cfg.AutoUpdate)
					// Save the updated config if needed
				case <-mAbout.ClickedCh:
					fmt.Println("Print Adapter v1.0.0\nDeveloped by Your Name")
				case <-mQuit.ClickedCh:
					systray.Quit()
				}
			}
		}()
	}
}

func onExit() {
	// Clean up here
	fmt.Println("Exiting application...")
}

func openFile(filePath string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", filePath).Start()
	case "windows":
		err = exec.Command("notepad", filePath).Start()
	case "darwin":
		err = exec.Command("open", filePath).Start()
	}
	if err != nil {
		fmt.Println("Failed to open file:", err)
	}
}

func openFolder(folderPath string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", folderPath).Start()
	case "windows":
		err = exec.Command("explorer", folderPath).Start()
	case "darwin":
		err = exec.Command("open", folderPath).Start()
	}
	if err != nil {
		fmt.Println("Failed to open folder:", err)
	}
}

//go:embed icon.png
var iconData []byte

func getIcon() []byte {
	return iconData
}
