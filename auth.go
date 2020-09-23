package main

import (
	"net/http"
)

func IsAuthorized(w http.ResponseWriter, r *http.Request, keyToCompare string) bool {
	auth := (*r).Header.Get("Authorization")
	if auth != keyToCompare {
		SendTextResponse(&w, "Not authorized.", http.StatusUnauthorized)
		return false
	} else {
		return true
	}
}