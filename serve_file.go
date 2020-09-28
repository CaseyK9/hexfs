package main

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

// ServeFile will serve the /{id} endpoint of hexFS. It gets the "id" variable from mux and tries to find the file's information in the database.
// If an ID is either not provided or not found, the function hands the request off to ServeNotFound.
func (b *BaseHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := mux.Vars(r)["id"]
	ext := path.Ext(id)

	if id == "" {
		ServeNotFound(w, r)
		return
	}

	f, e := b.GetFileData(FileData{ID: strings.TrimSuffix(id, ext), Ext: ext })
	if e != nil {
		SendTextResponse(&w, "Failed to get file information. " + e.Error(), http.StatusInternalServerError)
		return
	}
	if f == nil {
		ServeNotFound(w, r)
		return
	}

	wc, e := b.GCSClient.Bucket(os.Getenv(GCSBucketName)).Object(f.ID + f.Ext).Key(b.Key).NewReader(context.Background())
	if e != nil {
		if e == storage.ErrObjectNotExist {
			ServeNotFound(w, r)
			return
		}
		SendTextResponse(&w, "There was a problem reading the file. " + e.Error(), http.StatusInternalServerError)
		return
	}
	defer wc.Close()
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Type", wc.Attrs.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(wc.Attrs.Size, 10))
	_, copyErr := io.Copy(w, wc)
	if copyErr != nil {
		SendTextResponse(&w, "Could not write file to client. " + copyErr.Error(), http.StatusInternalServerError)
		return
	}
}
