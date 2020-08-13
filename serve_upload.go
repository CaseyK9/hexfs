package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

func ServeUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ApplyCORSHeaders(&w)

	maxSize, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	// Add 1024 bytes for headers, etc.
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+1024)

	parseErr := (*r).ParseMultipartForm(32<<20)
	if parseErr != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "File exceeds maximum size allowed.",
		})
		return
	}

	auth := IsAuthorized(r)
	if !auth {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Not authorized to upload.",
		})
		return
	}

	file, handler, err := (*r).FormFile("file") // Retrieve the file from form data
	if err != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "No files were uploaded.",
		})
		return
	}

	// Close form file
	defer func() {
		fileCloseErr := file.Close()
		if fileCloseErr != nil {
			fmt.Println("Couldn't close form file: " + fileCloseErr.Error())
		}
	}()
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
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "File is smaller than minimum size allowed.",
		})
		return
	}

	n, _ := strconv.ParseInt(os.Getenv(UploadDirMaxSize), 0, 64)
	if SizeOfUploadDir+ handler.Size > n {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Upload directory would exceed max size with this upload.",
		})
		return
	}
	var wg sync.WaitGroup
	randomStringChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		RandStringBytesMaskImprSrcUnsafe(8, randomStringChan, func() { wg.Done() })
	}()
	wg.Wait()

	fileId := <- randomStringChan

	f, openErr := os.OpenFile(path.Join(os.Getenv(UploadDirPath), fmt.Sprintf("%s%s", fileId, filepath.Ext(handler.Filename))), os.O_WRONLY|os.O_CREATE, 0666)
	if openErr != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "There was a problem trying to open the file. " + openErr.Error(),
		})
		return
	}
	// Close opened file on disk
	defer func() {
		fileCloseErr := f.Close()
		if fileCloseErr != nil {
			fmt.Println("Failed to close file on disk: " + fileCloseErr.Error())
		}
	}()
	written, writeErr := io.Copy(f, file)
	if writeErr != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "There was a problem writing the file to disk. " + writeErr.Error(),
		})
		return
	}

	SizeOfUploadDir += written
	fileName := fileId + filepath.Ext(handler.Filename)

	fullFileUrl := ""
	if os.Getenv(Endpoint) != "" {
		baseUrl, urlErr := url.Parse(os.Getenv(Endpoint))
		if urlErr != nil {
			SendJSONResponse(&w, ResponseError{
				Status:  1,
				Message: "Malformed endpoint. " + urlErr.Error(),
			})
			return
		}
		baseUrl.Path = path.Join(baseUrl.Path, fileName)
		fullFileUrl = baseUrl.String()
	}

	if os.Getenv(DiscordWebhookURL) != "" {
		sendStr := fileName
		if fullFileUrl != "" {
			sendStr = fullFileUrl
		}
		webhookErr := SendToWebhook(fmt.Sprintf("%s created. Wrote **%s** of data. (Total %% of space used: %.2f%%)", sendStr, ByteCountSI(uint64(written)), float64(SizeOfUploadDir) / float64(n)))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}

	SendJSONResponse(&w, UploadResponseSuccess{
		Status: 0,
		FileId: fileName,
		FullFileUrl: fullFileUrl,
		Size: handler.Size,
	})
}
