package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func getPCInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := map[string]any{
		"status":  1,
		"message": "OK",
		"data":    cachedPCInfo,
	}

	json.NewEncoder(w).Encode(response)
}

func Print(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, " @TODO PRINT")
}
