package main

import (
	"net/http"
	"os"
)

func ServePing(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	resText := "public mode disabled"
	if os.Getenv(PublicMode) == "1" {
		resText = "public mode enabled"
	}
	SendTextResponse(&w, resText, http.StatusOK)
	return
}

