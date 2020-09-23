package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

const IdLen = 8

func ServeUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if os.Getenv(PublicMode) != "1" {
		if !IsAuthorized(w, r, os.Getenv(UploadKey)) {
			return
		}
	}
	maxSize, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+1024)
	parseErr := (*r).ParseMultipartForm(32<<20)
	if parseErr != nil {
		SendTextResponse(&w, "File exceeds maximum size allowed/malformed body.", http.StatusBadRequest)
		return
	}

	file, handler, err := (*r).FormFile("file") // Retrieve the file from form data
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

	if os.Getenv(DisableFileBlacklist) == "0" {
		fileBlacklist := []string{".exe", ".com", ".dll", ".vbs", ".html", ".mhtml", ".xls", ".doc", ".xlsx", ".sh", ".bat", ".zsh", ""}
		for _, f := range fileBlacklist {
			if path.Ext(handler.Filename) == f {
				SendTextResponse(&w, "File extension prohibited.", http.StatusForbidden)
				return
			}
		}
	}

	if len(handler.Filename) > 128 {
		SendTextResponse(&w, "File name should not exceed 128 characters.", http.StatusBadRequest)
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

	var wg sync.WaitGroup
	randomStringChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		RandStringBytesMaskImprSrcUnsafe(IdLen, randomStringChan, func() { wg.Done() })
	}()
	wg.Wait()
	fileId := <- randomStringChan
	fileName := fileId + path.Ext(handler.Filename)

	key, err := GetKey()
	if err != nil {
		SendTextResponse(&w, "AES256 key not properly formatted to Base64.", http.StatusInternalServerError)
		return
	}
	wc := gcsClient.Bucket(os.Getenv(GCSBucketName)).Object(fileName).Key(key).NewWriter(ctx)
	defer wc.Close()
	written, writeErr := io.Copy(wc, file)
	if writeErr != nil {
		SendTextResponse(&w, "There was a problem writing the file to GCS. " + writeErr.Error(), http.StatusInternalServerError)
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
	baseUrl.Path = fileName
	fmt.Println(fmt.Sprintf("++++ %s | Size: %s | Original: `%s` | From IP: %s", fileName, ByteCountSI(uint64(written)), handler.Filename, GetIP(r)))
	SendTextResponse(&w, baseUrl.String(), http.StatusOK)
}