package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func ServePing(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	SendNothing(&w)
	return
}

