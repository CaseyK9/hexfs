package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/httputils/net"
	"github.com/vysiondev/httputils/rand"
	"io"
	"os"
	"path"
	"strconv"
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

	currentCap, err := b.RedisClient.Get(ctx, RedisKeyCurrentCapacity).Result()
	if err == redis.Nil {
		SendTextResponse(ctx, "Current capacity is unknown. For this reason, uploads are disabled.", fasthttp.StatusInternalServerError)
		return
	} else if err != nil {
		SendTextResponse(ctx, "Failed to determine the current capacity of the host. " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	maxCap, err := b.RedisClient.Get(ctx, RedisKeyMaxCapacity).Result()
	if err == redis.Nil {
		SendTextResponse(ctx, "Maximum capacity is unknown. For this reason, uploads are disabled.", fasthttp.StatusInternalServerError)
		return
	} else if err != nil {
		SendTextResponse(ctx, "Failed to determine the maximum capacity of the host. " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	currentCapParse, err := strconv.ParseInt(currentCap, 10, 64)
	if err != nil {
		SendTextResponse(ctx, "Failed to parse current capacity of the host. " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	maxCapParse, err := strconv.ParseInt(maxCap, 10, 64)
	if err != nil {
		SendTextResponse(ctx, "Failed to parse maximum capacity of the host. " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	if currentCapParse + f.Size > maxCapParse {
		SendTextResponse(ctx, "Host has reached its max capacity. No new files are allowed.", fasthttp.StatusBadRequest)
		return
	}

	if len(path.Ext(f.Filename)) > 20 {
		SendTextResponse(ctx, "File extension cannot be greater than 20 characters.", fasthttp.StatusBadRequest)
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
		rand.RandBytes(IdLen, randomStringChan, func() { wg.Done() })
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
		IP:                net.GetIP(ctx),
		Size:              written,
	})

	if e != nil {
		SendTextResponse(ctx, "Failed to insert new document. " + e.Error(), fasthttp.StatusInternalServerError)
		_, e := b.DeleteFiles(&FileData{ID:fileId})
		if e != nil {
			SendTextResponse(ctx, "Failed to delete file after failing to insert document. " + e.Error(), fasthttp.StatusInternalServerError)
		}
		return
	}

	e = b.RedisClient.Set(dbCtx, RedisKeyCurrentCapacity, currentCapParse + f.Size, 0).Err()
	if e != nil {
		SendTextResponse(ctx, "Failed to update capacity. " + e.Error(), fasthttp.StatusInternalServerError)
	}
	u := fmt.Sprintf("%s/%s", net.GetRoot(ctx), fileName)
	SendTextResponse(ctx, u, fasthttp.StatusOK)
}