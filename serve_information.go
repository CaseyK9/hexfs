package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"path"
	"strings"
)

func (b *BaseHandler) ServeInformation(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := mux.Vars(r)
	ext := path.Ext(id["id"])

	d, err := b.GetFileData(FileData{ID: strings.TrimSuffix(id["id"], ext), Ext: ext })
	if err != nil {
		SendTextResponse(&w, "Failed to fetch information. " + err.Error(), http.StatusInternalServerError)
		return
	}
	if d == nil {
		SendTextResponse(&w, "Not found.", http.StatusNotFound)
		return
	}
	// Redact IP if not using the master key.
	if GetAuthorizationLevel(r.Header.Get("authorization")) != IsMasterKey {
		d.IP = ""
	}
	SendJSONResponse(&w, d)
	return
}


