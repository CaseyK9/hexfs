package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

// ServeCheckAuth validates either the standard or master key.
func ServeCheckAuth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if GetAuthorizationLevel(mux.Vars(r)["key"]) == NotAuthorized {
		SendTextResponse(&w, "Not authorized.", http.StatusUnauthorized)
		return
	}
	SendNothing(&w)
}


