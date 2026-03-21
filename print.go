package main

import (
	"fmt"
	"net/http"
)

func pcInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, " @TODO PC-INFO")
}

func print(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, " @TODO PRINT")
}
