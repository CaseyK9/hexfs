package main

import (
	"net/http"
	"os"
)

func ServePing(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	SendTextResponse(&w, os.Getenv(PublicMode), http.StatusOK)
	return
}

