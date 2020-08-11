package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	VERSION = "1.0.0"
)

var sizeOfUploadDir int64

func HandleDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	applyCORSHeaders(&w)
	auth := IsAuthorized(r)
	if !auth {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Not authorized to delete.",
		})
		return
	}
	if ps.ByName("name") == "404.png" {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "This file is reserved and cannot be deleted.",
		})
		return
	}
	fi, statErr := os.Stat(path.Join(os.Getenv(UploadDirPath), ps.ByName("name")))
	if statErr != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Failed to get information about file. " + statErr.Error(),
		})
		return
	}
	err := os.Remove(path.Join(os.Getenv(UploadDirPath), ps.ByName("name")))
	if err != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Failed to delete file. " + err.Error(),
		})
		return
	}
	if os.Getenv(DiscordWebhookURL) != "" {
		webhookErr := SendToWebhook(fmt.Sprintf("%s deleted. Freed **%s** of space.", ps.ByName("name"), ByteCountSI(uint64(fi.Size()))))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}
	sizeOfUploadDir -= fi.Size()
	_ = json.NewEncoder(w).Encode(EmptyResponse{
		Status:  0,
	})
	return
}

func HandleUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	applyCORSHeaders(&w)
	auth := IsAuthorized(r)
	if !auth {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Not authorized to upload.",
		})
		return
	}

	contentTypeHeader := (*r).Header.Get("Content-Type")
	if !strings.Contains(contentTypeHeader, "multipart/form-data") {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Request's Content-Type must be multipart/form-data.",
		})
		return
	}

	maxSize, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	parseErr := (*r).ParseMultipartForm(maxSize)
	if parseErr != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "File exceeds maximum size allowed.",
		})
		return
	}

	file, handler, err := (*r).FormFile("file") // Retrieve the file from form data
	if err != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
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

	if handler.Size > maxSize {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "File exceeds maximum size allowed.",
		})
		return
	}

	minSize, _ := strconv.ParseInt(os.Getenv(MinSizeBytes), 0, 64)
	if handler.Size < minSize {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "File is smaller than minimum size allowed.",
		})
		return
	}

	n, _ := strconv.ParseInt(os.Getenv(UploadDirMaxSize), 0, 64)
	if sizeOfUploadDir+ handler.Size > n {
		_ = json.NewEncoder(w).Encode(ResponseError{
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
		_ = json.NewEncoder(w).Encode(ResponseError{
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
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "There was a problem writing the file to disk. " + writeErr.Error(),
		})
		return
	}

	sizeOfUploadDir += written
	fileName := fileId + filepath.Ext(handler.Filename)

	fullFileUrl := ""
	if os.Getenv(Endpoint) != "" {
		baseUrl, urlErr := url.Parse(os.Getenv(Endpoint))
		if urlErr != nil {
			_ = json.NewEncoder(w).Encode(ResponseError{
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
		webhookErr := SendToWebhook(fmt.Sprintf("%s created. Wrote **%s** of data. (Total %% of space used: %.2f%%)", sendStr, ByteCountSI(uint64(written)), float64(sizeOfUploadDir) / float64(n)))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}

	_ = json.NewEncoder(w).Encode(UploadResponseSuccess{
		Status: 0,
		FileId: fileName,
		FullFileUrl: fullFileUrl,
		Size: handler.Size,
	})
}

func ServeFileOrStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	applyCORSHeaders(&w)
	if (*r).Method == "OPTIONS" {
		return
	}
	if strings.ToLower(ps.ByName("name")) == "stats" {
		ServeStats(w, r, ps)
		return
	} else if strings.ToLower(ps.ByName("name")) == "ping" {
		_ = json.NewEncoder(w).Encode(EmptyResponse{
			Status:  0,
		})
		return
	}
	f, openErr := os.Open(path.Join(os.Getenv(UploadDirPath), ps.ByName("name")))
	if openErr != nil {
		notFoundImage, notFoundErr := os.Open(path.Join(os.Getenv(UploadDirPath), "404.png"))
		// Fallback if the image doesn't exist
		if notFoundErr != nil {
			_ = json.NewEncoder(w).Encode(ResponseError{
				Status:  1,
				Message: "File not found." + openErr.Error(),
			})
			return
		} else {
			defer func() {
				nfErr := notFoundImage.Close()
				if nfErr != nil {
					fmt.Println("Failed to close not-found image: " + nfErr.Error())
				}
			}()
			_, copyErr := io.Copy(w, notFoundImage)
			if copyErr != nil {
				_ = json.NewEncoder(w).Encode(ResponseError{
					Status:  1,
					Message: "Could not write 404 image to client.",
				})
				return
			}
			return
		}
	}
	defer func() {
		fileErr := f.Close()
		if fileErr != nil {
			fmt.Println("Failed to close image sent to client: " + fileErr.Error())
		}
	}()
	header := make([]byte, 512)
	_, e := f.Read(header)
	if e != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Could not read file headers.",
		})
		return
	}
	contentType := http.DetectContentType(header)
	fileStat, _ := f.Stat()
	size := strconv.FormatInt(fileStat.Size(), 10)

	_, _ = f.Seek(0, 0)
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", size)
	_, copyErr := io.Copy(w, f)
	if copyErr != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Could not write file to client.",
		})
		return
	}
	return
}
func ServeStats(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	applyCORSHeaders(&w)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	n, _ := strconv.ParseInt(os.Getenv(UploadDirMaxSize), 0, 64)
	max, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	min, _ := strconv.ParseInt(os.Getenv(MinSizeBytes), 0, 64)

	_ = json.NewEncoder(w).Encode(StatsResponseSuccess{
		Status:         0,
		WebhookEnabled: os.Getenv(DiscordWebhookURL) != "",
		MemAllocated:   ByteCountSI(mem.Alloc),
		MaxFileSize:    ByteCountSI(uint64(max)),
		MinFileSize:    ByteCountSI(uint64(min)),
		SpaceMax:       n,
		SpaceUsed:      sizeOfUploadDir,
		Version:        VERSION,
	})
	return
}

func applyCORSHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
}

func main() {
	envErr := godotenv.Load()
	if envErr != nil {
		panic("Cannot find a .env file in the project root.")
	}
	ValidateEnv()
	if _, err := os.Stat(os.Getenv(UploadDirPath)); err != nil {
		if os.IsNotExist(err) {
			panic("Directory " + os.Getenv(UploadDirPath) + " does not exist. Create it and try again.")
		}
	}
	s, e := DirSize(os.Getenv(UploadDirPath))
	if e != nil {
		panic(e)
	}
	sizeOfUploadDir = s
	router := httprouter.New()
	router.GET("/:name", ServeFileOrStats)
	router.POST("/", HandleUpload)
	router.POST("/delete/:name", HandleDelete)
	fmt.Println("Listening on " + os.Getenv(Port))
	log.Fatal(http.ListenAndServe(":" + os.Getenv(Port), router))
}
