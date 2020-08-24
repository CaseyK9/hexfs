package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

func ServeIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch strings.ToLower(ps.ByName("id")) {
	case "ping":
		ServePing(w, r, ps)
		break
	default:
		NotFoundHandler(w, r)
	}
}