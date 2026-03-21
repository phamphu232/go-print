package main

import (
	"fmt"
	"net/http"
)

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
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

	mux.HandleFunc("/pc-info", pcInfo)
	mux.HandleFunc("/print", print)

	return mux
}
