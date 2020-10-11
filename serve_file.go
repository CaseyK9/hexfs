// Parts of this file were derived from https://github.com/whats-this/cdn-origin/blob/8b05fa8425db01cce519ca8945203f9d3050c33b/main.go#L439.
// The implementation reason was a workaround found by this repository to prevent discord from hiding image URLs.

package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/httputils/net"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const rawParam = "raw"
const discordHTML = `<html>
	<head>
		<meta property="twitter:card" content="summary_large_image" />
		<meta property="twitter:image" content="{{.}}" />
		<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
		<meta http-equiv="Pragma" content="no-cache" />
		<meta http-equiv="Expires" content="0" />
	</head>
</html>`

var (
	discordBotRegex = regexp.MustCompile("(?i)discordbot")
)

// ServeFile will serve the / endpoint of hexFS. It gets the "id" variable from mux and tries to find the file's information in the database.
// If an ID is either not provided or not found, the function hands the request off to ServeNotFound.
func (b *BaseHandler) ServeFile(ctx *fasthttp.RequestCtx) {
	id := ctx.Request.URI().LastPathSegment()
	if len(id) == 0 {
		ServeNotFound(ctx)
		return
	}
	ext := path.Ext(string(id))

	f, e := b.GetFileData(FileData{ID: strings.TrimSuffix(string(id), ext), Ext: ext })
	if e != nil {
		SendTextResponse(ctx, "Failed to get file information. " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}
	if f == nil {
		SendTextResponse(ctx, "Not found.", fasthttp.StatusNotFound)
		return
	}

	fileCtx, cancel := context.WithTimeout(context.Background(), time.Minute * 10)
	defer cancel()

	wc, e := b.GCSClient.Bucket(os.Getenv(GCSBucketName)).Object(f.ID + f.Ext).Key(b.Key).NewReader(fileCtx)
	if e != nil {
		if e == storage.ErrObjectNotExist {
			ServeNotFound(ctx)
			return
		}
		SendTextResponse(ctx, "There was a problem reading the file. " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}
	defer wc.Close()

	if discordBotRegex.Match(ctx.Request.Header.UserAgent()) && !ctx.QueryArgs().Has(rawParam) {
		if wc.Attrs.ContentType == "image/png" || wc.Attrs.ContentType == "image/jpeg" || wc.Attrs.ContentType == "image/gif" || wc.Attrs.ContentType == "image/apng" {
			ctx.Response.Header.SetContentType("text/html; charset=utf8")
			ctx.Response.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
			ctx.Response.Header.Add("Pragma", "no-cache")
			ctx.Response.Header.Add("Expires", "0")
		}
		url := fmt.Sprintf("%s/%s?%s=true", net.GetRoot(ctx), f.ID + f.Ext, rawParam)
		_, _ = fmt.Fprint(ctx.Response.BodyWriter(), strings.Replace(discordHTML, "{{.}}", url, 1))
		return
	}

	ctx.Response.Header.Add("Content-Disposition", "inline")
	ctx.Response.Header.Add("Content-Length", strconv.FormatInt(wc.Attrs.Size, 10))
	ctx.Response.Header.Add("Content-Type", wc.Attrs.ContentType)
	_, copyErr := io.Copy(ctx.Response.BodyWriter(), wc)
	if copyErr != nil {
		SendTextResponse(ctx, "Could not write file to client. " + copyErr.Error(), fasthttp.StatusInternalServerError)
		return
	}
}
