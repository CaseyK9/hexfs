package main

import (
	"github.com/valyala/fasthttp"
	"path"
	"strings"
)

func (b *BaseHandler) ServeInformation(ctx *fasthttp.RequestCtx) {
	id := ctx.QueryArgs().Peek("id")
	if len(id) == 0 {
		SendTextResponse(ctx, "No ID to search given. ", fasthttp.StatusBadRequest)
		return
	}
	ext := path.Ext(string(id))

	d, err := b.GetFileData(ctx, FileData{ID: strings.TrimSuffix(string(id), ext), Ext: ext })
	if err != nil {
		SendTextResponse(ctx, "Failed to fetch information. " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	if d == nil {
		SendTextResponse(ctx, "Not found.", fasthttp.StatusNotFound)
		return
	}
	// Redact IP if not using the master key.
	if b.GetAuthorizationLevel(ctx.Request.Header.Peek("Authorization")) != IsMasterKey {
		d.IP = ""
	}
	SendJSONResponse(ctx, d)
	return
}


