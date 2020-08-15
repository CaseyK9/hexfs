package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

func ServeIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if (*r).Method == "OPTIONS" {
		return
	}
	switch strings.ToLower(ps.ByName("id")) {
	case "stats":
		ServeStats(w, r, ps)
		break
	case "ping":
		ServePing(w, r, ps)
		break
	default:
		NotFoundHandler(w, r)
	}
}