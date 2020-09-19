package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"time"
)

func ServeDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gcsClient, ctx, e := CreateGCSClient()
	if e != nil {
		SendTextResponse(&w, "There was a problem creating the GCS Client. " + e.Error(), http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	defer gcsClient.Close()

	key, err := GetKey()
	if err != nil {
		SendTextResponse(&w, "AES256 key not properly formatted to Base64.", http.StatusInternalServerError)
		return
	}

	f := gcsClient.Bucket(os.Getenv(GCSBucketName)).Object(ps.ByName("name")).Key(key)
	s, err := f.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			SendTextResponse(&w, "File does not exist.", http.StatusNotFound)
			return
		} else {
			SendTextResponse(&w, "Failed to read file info from GCS. " + err.Error(), http.StatusInternalServerError)
			return
		}
	}
	defer s.Close()
	err = f.Delete(ctx)
	if err != nil {
		SendTextResponse(&w, "Failed to remove file from GCS. " + err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(fmt.Sprintf("---- `%s` | Size: %s", ps.ByName("name"), ByteCountSI(uint64(s.Attrs.Size))))
	SendNothing(&w)
	return
}
