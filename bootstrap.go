package main

func bootstrap() {
	loadSetting()
	initLogger()
	setCachePCInfo()
	makeDownloadDir()
	startCleanupWorker()
}
