package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/httputils/net"
	"github.com/vysiondev/httputils/rand"
	"io"
	"path"
	"sync"
)

const fileHandler = "file"

// ServeUpload handles all incoming POST requests to /upload. It will take a multipart form, parse the file, then write it to both GCS and a hasher at the same time.
// The file's information will also be inserted into the database.
func (b *BaseHandler) ServeUpload(ctx *fasthttp.RequestCtx) {
	auth := b.IsAuthorized(ctx)
	if !auth && !b.Config.Security.PublicMode {
		SendTextResponse(ctx, "Not authorized to upload.", fasthttp.StatusUnauthorized)
		return
	}

	mp, e := ctx.Request.MultipartForm()
	if e != nil {
		if e == fasthttp.ErrNoMultipartForm {
			SendTextResponse(ctx, "Multipart form not sent.", fasthttp.StatusBadRequest)
			return
		}
		SendTextResponse(ctx, "There was a problem parsing the form. " + e.Error(), fasthttp.StatusBadRequest)
		return
	}
	if len(mp.File[fileHandler]) == 0 {
		SendTextResponse(ctx, "No files were uploaded.", fasthttp.StatusBadRequest)
	}
	f := mp.File[fileHandler][0]

	if len(path.Ext(f.Filename)) > 20 {
		SendTextResponse(ctx, "File extension cannot be greater than 20 characters.", fasthttp.StatusBadRequest)
		return
	}

	if len(f.Filename) > 256 {
		SendTextResponse(ctx, "File name should not exceed 256 characters.", fasthttp.StatusBadRequest)
		return
	}
	ext := path.Ext(f.Filename)
	if len(ext) == 0 {
		SendTextResponse(ctx, "Files with no extension prohibited.", fasthttp.StatusBadRequest)
		return
	}
	if len(b.Config.Security.Blacklist) > 0 {
		for _, t := range b.Config.Security.Blacklist {
			if ext == "." + t {
				SendTextResponse(ctx, "Extension blacklisted.", fasthttp.StatusBadRequest)
				return
			}
		}
	}
	if len(b.Config.Security.Whitelist) > 0 {
		for i, t := range b.Config.Security.Whitelist {
			if ext == "." + t {
				break
			}
			if i + 1 == len(b.Config.Security.Whitelist) {
				SendTextResponse(ctx, "Extension not whitelisted.", fasthttp.StatusBadRequest)
				return
			}
		}
	}

	var wg sync.WaitGroup
	randomStringChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		rand.RandBytes(b.Config.Server.IDLen, randomStringChan, func() { wg.Done() })
	}()
	wg.Wait()
	fileId := <- randomStringChan
	fileName := fileId + path.Ext(f.Filename)

	wc := b.GCSClient.Bucket(b.Config.Net.GCS.BucketName).Object(fileName).Key(b.Key).NewWriter(ctx)
	defer wc.Close()

	openedFile, e := f.Open()
	if e != nil {
		SendTextResponse(ctx, "Failed to open file from request: " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer openedFile.Close()

	_, writeErr := io.Copy(wc, openedFile)
	if writeErr != nil {
		SendTextResponse(ctx, "There was a problem writing the file to GCS. " + writeErr.Error(), fasthttp.StatusInternalServerError)
		return
	}

	u := fmt.Sprintf("%s/%s", net.GetRoot(ctx), fileName)
	SendTextResponse(ctx, u, fasthttp.StatusOK)
}