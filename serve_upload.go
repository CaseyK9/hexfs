package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

const IdLen = 9
const fileHandler = "file"

// ServeUpload handles all incoming POST requests to /. It will take a multipart form, parse the file, then write it to both GCS and a hasher at the same time.
// The file's information will also be inserted into the database.
func (b *BaseHandler) ServeUpload(ctx *fasthttp.RequestCtx) {
	auth := GetAuthorizationLevel(ctx.Request.Header.Peek("Authorization"))
	if auth == NotAuthorized && os.Getenv(PublicMode) != "1" {
		SendTextResponse(ctx, "Not authorized to upload.", fasthttp.StatusUnauthorized)
		return
	}

	mp, e := ctx.Request.MultipartForm()
	if e != nil {
		if e == fasthttp.ErrNoMultipartForm {
			SendTextResponse(ctx, "Multipart form not sent.", fasthttp.StatusBadRequest)
			return
		}
		SendTextResponse(ctx, "Multipart form not sent.", fasthttp.StatusBadRequest)
		return
	}
	if len(mp.File[fileHandler]) == 0 {
		SendTextResponse(ctx, "No files were uploaded.", fasthttp.StatusBadRequest)
	}
	f := mp.File[fileHandler][0]

	if len(path.Ext(f.Filename)) > 20 {
		SendTextResponse(ctx, "File extension cannot be greater than 12 characters.", fasthttp.StatusBadRequest)
		return
	}

	if len(f.Filename) > 256 {
		SendTextResponse(ctx, "File name should not exceed 256 characters.", fasthttp.StatusBadRequest)
		return
	}
	if os.Getenv(DisableFileBlacklist) == "0" {
		fileBlacklist := []string{".exe", ".com", ".dll", ".vbs", ".html", ".mhtml", ".xls", ".doc", ".xlsx", ".sh", ".bat", ".zsh", ""}
		for _, t := range fileBlacklist {
			if path.Ext(f.Filename) == t {
				SendTextResponse(ctx, "File extension prohibited.", fasthttp.StatusBadRequest)
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
	fileName := fileId + path.Ext(f.Filename)

	writerCtx, cancel := context.WithTimeout(context.Background(), time.Minute * 5)
	defer cancel()

	wc := b.GCSClient.Bucket(os.Getenv(GCSBucketName)).Object(fileName).Key(b.Key).NewWriter(writerCtx)
	defer wc.Close()

	hasher := sha256.New()

	openedFile, e := f.Open()
	if e != nil {
		SendTextResponse(ctx, "Failed to open file from request: " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer openedFile.Close()

	written, writeErr := io.Copy(io.MultiWriter(wc, hasher), openedFile)
	if writeErr != nil {
		SendTextResponse(ctx, "There was a problem writing the file to GCS and/or detecting the SHA256 signature. " + writeErr.Error(), fasthttp.StatusInternalServerError)
		return
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), time.Second * 30)
	defer cancel()

	_, e = b.Database.Collection(MongoCollectionFiles).InsertOne(dbCtx, &FileData{
		ID:                fileId,
		Ext:               path.Ext(f.Filename),
		SHA256:            hex.EncodeToString(hasher.Sum(nil)),
		UploadedTimestamp: time.Now().Format(time.RFC3339),
		IP:                GetIP(ctx),
		Size:              written,
	})

	if e != nil {
		SendTextResponse(ctx, "Failed to insert new document. " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}

	u := fmt.Sprintf("%s/%s", GetRoot(ctx), fileName)
	SendTextResponse(ctx, u, fasthttp.StatusOK)
}