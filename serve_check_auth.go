package main

import (
	"net/http"
)

// ServeCheckAuth validates either the standard or master key.
func ServeCheckAuth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if GetAuthorizationLevel(r.Header.Get("authorization")) == NotAuthorized {
		SendTextResponse(&w, "Not authorized.", http.StatusUnauthorized)
		return
	}
	SendNothing(&w)
}


