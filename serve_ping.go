package main

import (
	"github.com/valyala/fasthttp"
)

func (b *BaseHandler)ServePing(ctx *fasthttp.RequestCtx) {
	resText := "public mode disabled"
	if b.Config.Security.PublicMode {
		resText = "public mode enabled"
	}
	SendTextResponse(ctx, resText, fasthttp.StatusOK)
}

