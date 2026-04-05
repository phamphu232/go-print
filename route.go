package main

import (
	"embed"
	"fmt"
	"net/http"
)

//go:embed favicon.ico
var faviconFile embed.FS

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, _ := faviconFile.ReadFile("favicon.ico")
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(data)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newURL := fmt.Sprintf("http://%s:%d/setting/", setting.Host, setting.Port)
		http.Redirect(w, r, newURL, http.StatusFound)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	mux.HandleFunc("/setting/", editSetting)
	mux.HandleFunc("/setting/update", updateSetting)
	mux.HandleFunc("/pc-info", getPCInfo)
	mux.HandleFunc("/printer-info", getPrinterInfo)
	mux.HandleFunc("/print", Print)
	mux.HandleFunc("/service/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, statusService())
	})

	return mux
}
