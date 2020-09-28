package main

import (
	"net/http"
	"os"
)

func ServeNotFound(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if os.Getenv(Frontend) != "" {
		http.Redirect(w, r, os.Getenv(Frontend), http.StatusPermanentRedirect)
	} else {
		SendTextResponse(&w, "Page not found.", http.StatusNotFound)
	}
}
