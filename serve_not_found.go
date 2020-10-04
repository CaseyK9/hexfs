package main

import (
	"net/http"
)

func ServeNotFound(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	SendTextResponse(&w, "Page not found.", http.StatusNotFound)
}
