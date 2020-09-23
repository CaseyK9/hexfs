package main

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func ServeNotFound(w http.ResponseWriter, r *http.Request) {
	if os.Getenv(Frontend) != "" {
		http.Redirect(w, r, os.Getenv(Frontend), http.StatusPermanentRedirect)
	} else {
		SendTextResponse(&w, "Page not found.", http.StatusNotFound)
	}
}
func ServeFileOrNotFound(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("id") == "" {
		ServeNotFound(w, r)
		return
	}
	gcsClient, e := CreateGCSClient()
	if e != nil {
		SendTextResponse(&w, "There was a problem creating the GCS Client. " + e.Error(), http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	defer gcsClient.Close()
	key, err := GetKey()
	if err != nil {
		SendTextResponse(&w, "AES256 key not properly formatted to Base64.", http.StatusInternalServerError)
		return
	}
	wc, e := gcsClient.Bucket(os.Getenv(GCSBucketName)).Object(ps.ByName("id")).Key(key).NewReader(ctx)
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
