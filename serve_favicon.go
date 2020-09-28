package main

import (
	"io"
	"net/http"
	"os"
)

func ServeFavicon(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	f, e := os.OpenFile("favicon.ico", os.O_RDONLY, 0666)
	if e != nil {
		if e == os.ErrNotExist {
			w.WriteHeader(204)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	defer f.Close()
	_, e = io.Copy(w, f)
	if e != nil {
		w.WriteHeader(500)
	}
}
