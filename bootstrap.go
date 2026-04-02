package main

func bootstrap() {
	initLogger()
	loadSetting()
	setCachePCInfo()
	makeDownloadDir()
	startCleanupWorker()
}
