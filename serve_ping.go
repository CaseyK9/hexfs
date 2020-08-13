package main

import (
	"net/http"
)

func ServePing(w http.ResponseWriter) {
	SendJSONResponse(&w, EmptyResponse{
		Status:  0,
	})
	return
}

