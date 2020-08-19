package main

import (
	"net/http"
	"os"
)

func IsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	auth := (*r).Header.Get("Authorization")
	if auth != os.Getenv(UploadKey) {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Not authorized.",
		}, http.StatusUnauthorized)
		return false
	} else {
		return true
	}
}
