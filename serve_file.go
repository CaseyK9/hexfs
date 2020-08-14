package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

func ServeFile(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	f, openErr := os.Open(path.Join(os.Getenv(UploadDirPath), ps.ByName("id"), ps.ByName("name")))
	if openErr != nil {
		Write404ToResponse(path.Join(os.Getenv(UploadDirPath), "../404.png"), w)
		return
	}
	defer func() {
		fileErr := f.Close()
		if fileErr != nil {
			fmt.Println("Failed to close image sent to client. " + fileErr.Error())
		}
	}()
	header := make([]byte, 512)
	_, e := f.Read(header)
	if e != nil {
		SendJSONResponse(&w, ResponseError{
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
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Could not write file to client.",
		})
		return
	}
	return
}

func Write404ToResponse(filePath string, w http.ResponseWriter) {
	f, openErr := os.Open(filePath)
	if openErr != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "File not found.",
		})
		return
	}
	defer func() {
		_ = f.Close()
	}()
	_, copyErr := io.Copy(w, f)
	if copyErr != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Could not write 404 image to client.",
		})
	}
}
