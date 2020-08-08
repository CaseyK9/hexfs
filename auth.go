package main

import (
	"net/http"
	"os"
)

func IsAuthorized(r *http.Request) bool {
	auth := (*r).Header.Get("Authorization")
	return auth == os.Getenv(UploadKey)
}
