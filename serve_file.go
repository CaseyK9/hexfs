// Parts of this file were derived from https://github.com/whats-this/cdn-origin/blob/8b05fa8425db01cce519ca8945203f9d3050c33b/main.go#L439.
// The implementation reason was a workaround found by this repository to prevent discord from hiding image URLs.

package main

import (
	"cloud.google.com/go/storage"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/httputils/net"
	"io"
	"regexp"
	"strconv"
	"strings"
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
		b.ServeNotFound(ctx)
		return
	}

	wc, e := b.GCSClient.Bucket(b.Config.Net.GCS.BucketName).Object(string(id)).Key(b.Key).NewReader(ctx)
	if e != nil {
		if e == storage.ErrObjectNotExist {
			b.ServeNotFound(ctx)
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
		url := fmt.Sprintf("%s/%s?%s=true", net.GetRoot(ctx), id, rawParam)
		_, _ = fmt.Fprint(ctx.Response.BodyWriter(), strings.Replace(discordHTML, "{{.}}", url, 1))
		return
	}
	filterStatus := b.FilterCheck(ctx, wc.Attrs.ContentType)
	if filterStatus == FilterFail {
		return
	} else if filterStatus == FilterSanitize {
		ctx.Response.Header.Set("Content-Type", "text/plain")
	} else {
		ctx.Response.Header.Set("Content-Type", wc.Attrs.ContentType)
	}
	ctx.Response.Header.Set("Content-Disposition", "inline")
	ctx.Response.Header.Set("Content-Length", strconv.FormatInt(wc.Attrs.Size, 10))
	_, copyErr := io.Copy(ctx.Response.BodyWriter(), wc)
	if copyErr != nil {
		SendTextResponse(ctx, "Could not write file to client. " + copyErr.Error(), fasthttp.StatusInternalServerError)
		return
	}
}
