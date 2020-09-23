package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
)

func ServePing(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()
	SendTextResponse(&w, os.Getenv(PublicMode), http.StatusOK)
	return
}

