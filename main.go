package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var SizeOfUploadDir int64

func HandleUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	auth := r.Header.Get("Authorization")
	if auth == "" || auth != os.Getenv(UploadKey) {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "Not authorized.",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	maxSize, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	parseErr := r.ParseMultipartForm(maxSize)
	if parseErr != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "File exceeds maximum size allowed.",
		})
		return
	}


	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "No files were uploaded.",
		})
		return
	}
	defer file.Close()

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
	if SizeOfUploadDir + handler.Size > n {
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
	defer f.Close()
	written, writeErr := io.Copy(f, file)
	if writeErr != nil {
		_ = json.NewEncoder(w).Encode(ResponseError{
			Status:  1,
			Message: "There was a problem writing the file to disk. " + writeErr.Error(),
		})
		return
	}

	SizeOfUploadDir += written

	if os.Getenv(DiscordWebhookURL) != "" {
		webhookErr := SendToWebhook(fmt.Sprintf("%s uploaded. Space used: %.2f%%", fileId + filepath.Ext(handler.Filename), float64(SizeOfUploadDir) / float64(n)))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}

	_ = json.NewEncoder(w).Encode(UploadResponseSuccess{
		Status: 0,
		FileId: fileId + filepath.Ext(handler.Filename),
		Size:   handler.Size,
	})
}
func ServeFileOrStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	applyCORSHeaders(&w)
	if strings.ToLower(ps.ByName("name")) == "stats" {
		ServeStats(w, r, ps)
		return
	} else if strings.ToLower(ps.ByName("name")) == "ping" {
		_ = json.NewEncoder(w).Encode(Ping{
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
			defer notFoundImage.Close()
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
	defer f.Close()
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
	w.Header().Set("Content-Type", "application/json")
	n, _ := strconv.ParseInt(os.Getenv(UploadDirMaxSize), 0, 64)
	max, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	min, _ := strconv.ParseInt(os.Getenv(MinSizeBytes), 0, 64)

	_ = json.NewEncoder(w).Encode(StatsResponseSuccess{
		Status:         0,
		WebhookEnabled: os.Getenv(DiscordWebhookURL) != "",
		MemAllocated:   ByteCountSI(mem.Alloc),
		MaxFileSize:    ByteCountSI(uint64(max)),
		MinFileSize:    ByteCountSI(uint64(min)),
		SpaceMax: 			n,
		SpaceUsed:      SizeOfUploadDir,
	})
	return
}

func applyCORSHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
}

func main() {
	ValidateEnv()
	s, e := DirSize(os.Getenv(UploadDirPath))
	if e != nil {
		panic(e)
	}
	SizeOfUploadDir = s
	router := httprouter.New()
	router.GET("/:name", ServeFileOrStats)
	router.POST("/", HandleUpload)
	fmt.Println("Listening on " + os.Getenv(Port))
	log.Fatal(http.ListenAndServe(":" + os.Getenv(Port), router))
}
