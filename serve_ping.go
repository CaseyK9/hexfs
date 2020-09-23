package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
)

func ServePing(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	SendTextResponse(&w, os.Getenv(PublicMode), http.StatusOK)
	return
}

