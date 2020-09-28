package main

import (
	"net/http"
	"os"
)

const (
	IsStandardKey = 0
	IsMasterKey = 1
	NotAuthorized = 2
)

func (b *BaseHandler) ProtectedRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if GetAuthorizationLevel(request.Header.Get("authorization")) != IsMasterKey {
			SendTextResponse(&writer, "Not authorized.", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(writer, request)
	}
}

func GetAuthorizationLevel(test string) int {
	switch test {
	case os.Getenv(MasterKey):
		return IsMasterKey
	case os.Getenv(StandardKey):
		return IsStandardKey
	default:
		return NotAuthorized
	}
}