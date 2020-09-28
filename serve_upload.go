package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
)

const IdLen = 9

// ServeUpload handles all incoming POST requests to /. It will take a multipart form, parse the file, then write it to both GCS and a hasher at the same time.
// The file's information will also be inserted into the database.
func (b *BaseHandler) ServeUpload(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, b.MaxSizeBytes + 1024)
	parseErr := (*r).ParseMultipartForm(32 << 20)
	if parseErr != nil {
		if parseErr == io.ErrUnexpectedEOF {
			SendTextResponse(&w, "File exceeds maximum size allowed.", http.StatusBadRequest)
		} else {
			SendTextResponse(&w, "Wrong or malformed body sent.", http.StatusBadRequest)
		}
		return
	}

	urlToSend := os.Getenv(Endpoint)
	if r.FormValue("proxy") != "" {
		urlToSend = r.FormValue("proxy")
	}
	baseUrl, urlErr := url.Parse(urlToSend)
	if urlErr != nil {
		SendTextResponse(&w, "Malformed endpoint. " + urlErr.Error(), http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		SendTextResponse(&w, "No files were uploaded.", http.StatusBadRequest)
		return
	}

	// Close form file
	defer func() {
		fileCloseErr := file.Close()
		if fileCloseErr != nil {
			fmt.Println("Couldn't close form file: " + fileCloseErr.Error())
		}
	}()

	if len(path.Ext(handler.Filename)) > 20 {
		SendTextResponse(&w, "File extension cannot be greater than 20 characters.", http.StatusBadRequest)
		return
	}

	if len(handler.Filename) > 128 {
		SendTextResponse(&w, "File name should not exceed 128 characters.", http.StatusBadRequest)
		return
	}
	if os.Getenv(DisableFileBlacklist) == "0" {
		fileBlacklist := []string{".exe", ".com", ".dll", ".vbs", ".html", ".mhtml", ".xls", ".doc", ".xlsx", ".sh", ".bat", ".zsh", ""}
		for _, f := range fileBlacklist {
			if path.Ext(handler.Filename) == f {
				SendTextResponse(&w, "File extension prohibited.", http.StatusForbidden)
				return
			}
		}
	}

	var wg sync.WaitGroup
	randomStringChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		RandStringBytesMaskImprSrcUnsafe(IdLen, randomStringChan, func() { wg.Done() })
	}()
	wg.Wait()
	fileId := <- randomStringChan
	fileName := fileId + path.Ext(handler.Filename)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()

	wc := b.GCSClient.Bucket(os.Getenv(GCSBucketName)).Object(fileName).Key(b.Key).NewWriter(ctx)
	defer wc.Close()

	hasher := sha256.New()

	written, writeErr := io.Copy(io.MultiWriter(wc, hasher), file)
	if writeErr != nil {
		SendTextResponse(&w, "There was a problem writing the file to GCS and/or detecting the SHA256 signature. " + writeErr.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()

	_, e := b.Database.Collection(MongoCollectionFiles).InsertOne(ctx, &FileData{
		ID:                fileId,
		Ext:               path.Ext(handler.Filename),
		SHA256:            hex.EncodeToString(hasher.Sum(nil)),
		UploadedTimestamp: time.Now().Format(time.RFC3339),
		IP:                GetIP(r),
		Size:              written,
	})

	if e != nil {
		SendTextResponse(&w, "Failed to insert new document. " + e.Error(), http.StatusInternalServerError)
		return
	}

	baseUrl.Path = fileName
	SendTextResponse(&w, baseUrl.String(), http.StatusOK)
}