package main

import (
	"net/http"
	"os"
)

func ServeNotFound(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.URL.Path == "/" && r.Method == http.MethodGet && len(os.Getenv(Frontend)) != 0 {
		http.Redirect(w, r, os.Getenv(Frontend), http.StatusMovedPermanently)
		return
	}
	SendTextResponse(&w, "Page not found.", http.StatusNotFound)
}
