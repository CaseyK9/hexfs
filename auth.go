package main

import (
	"net/http"
	"os"
)

func IsAuthorized(r *http.Request) *ResponseError {
	auth := (*r).Header.Get("Authorization")
	if auth != os.Getenv(UploadKey) {
		return &ResponseError{
			Status:  1,
			Message: "Not authorized.",
		}
	} else {
		return nil
	}
}
