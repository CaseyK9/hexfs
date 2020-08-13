package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

func ServeIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ApplyCORSHeaders(&w)
	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch strings.ToLower(ps.ByName("name")) {
	case "stats":
		ServeStats(w, r, ps)
		break
	case "ping":
		ServePing(w)
		break
	default:
		ServeFile(w, ps)
	}
}