package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
)

const IdLen = 6

func ServeUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	maxSize, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	authorized := true
	// Don't accept a large request body if the user is unauthorized
	if os.Getenv(UploadKey) != r.Header.Get("Authorization") {
		maxSize = 2048
		authorized = false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+1024)

	parseErr := (*r).ParseMultipartForm(32<<20)
	if parseErr != nil {
		if !authorized {
			SendTextResponse(&w, "Not authorized to upload.", http.StatusUnauthorized)
		} else {
			SendTextResponse(&w, "File exceeds maximum size allowed/malformed body.", http.StatusBadRequest)
		}
		return
	}
	if !HasContentType(r, "multipart/form-data") {
		SendTextResponse(&w, "Content-Type must equal multipart/form-data.", http.StatusBadRequest)
		return
	}

	authSuccess := IsAuthorized(w, r)
	if !authSuccess {
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
	if len(handler.Filename) > 64 {
		SendTextResponse(&w, "File name should not exceed 64 characters.", http.StatusBadRequest)
		return
	}
	// Not necessary?
	//if handler.Size > maxSize {
	//	_ = json.NewEncoder(w).Encode(ResponseError{
	//		Status:  1,
	//		Message: "File exceeds maximum size allowed.",
	//	})
	//	return
	//}

	minSize, _ := strconv.ParseInt(os.Getenv(MinSizeBytes), 0, 64)
	if handler.Size < minSize {
		SendTextResponse(&w, "File is smaller than minimum size allowed.", http.StatusBadRequest)
		return
	}

	n, _ := strconv.ParseInt(os.Getenv(UploadDirMaxSize), 0, 64)
	sizeOfUploadDir, err := DirSize(os.Getenv(UploadDirPath))
	if err != nil {
		SendTextResponse(&w, "Could not get upload directory size.", http.StatusInternalServerError)
		return
	}
	if sizeOfUploadDir + handler.Size > n {
		SendTextResponse(&w, "Upload directory would exceed max set size with this upload.", http.StatusInsufficientStorage)
		return
	}
	var wg sync.WaitGroup
	randomStringChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		RandStringBytesMaskImprSrcUnsafe(IdLen, randomStringChan, func() { wg.Done() })
	}()
	wg.Wait()

	fileId := <- randomStringChan
	filePath := path.Join(os.Getenv(UploadDirPath), fileId)
	mkdirErr := os.Mkdir(filePath, 0755)
	if mkdirErr != nil {
		SendTextResponse(&w, "Could not create directory for the file. ", http.StatusInternalServerError)
		return
	}

	f, openErr := os.OpenFile(path.Join(filePath, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if openErr != nil {
		rmErr := os.RemoveAll(filePath)
		if rmErr != nil {
			SendTextResponse(&w, "Failed to remove directory after failing to open file. RemoveAll err: " + rmErr.Error() + " OpenFile err: " + openErr.Error(), http.StatusInternalServerError)
			return
		}
		SendTextResponse(&w, "There was a problem trying to open the file. " + openErr.Error(), http.StatusInternalServerError)
		return
	}
	// Close opened file on disk. Never evaluates an error.
	defer func() {
		_ = f.Close()
	}()

	written, writeErr := io.Copy(f, file)
	if writeErr != nil {
		rmErr := os.RemoveAll(filePath)
		if rmErr != nil {
			SendTextResponse(&w, "Failed to remove directory after failing writing file to disk. RemoveAll err: " + rmErr.Error() + " Copy err: " + writeErr.Error(), http.StatusInternalServerError)
			return
		}
		SendTextResponse(&w, "There was a problem writing the file to disk. " + writeErr.Error(), http.StatusInternalServerError)
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
	baseUrl.Path = path.Join(fileId, handler.Filename)

	if os.Getenv(DiscordWebhookURL) != "" {
		webhookErr := SendToWebhook(fmt.Sprintf("%s created. Wrote **%s** of data. (Total %% of space used: %.2f%%)", baseUrl.String(), ByteCountSI(uint64(written)), float64(sizeOfUploadDir) / float64(n)))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}

	SendTextResponse(&w, baseUrl.String(), http.StatusOK)
}
