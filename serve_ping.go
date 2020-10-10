package main

import (
	"github.com/valyala/fasthttp"
	"os"
)

func ServePing(ctx *fasthttp.RequestCtx) {
	resText := "public mode disabled"
	if os.Getenv(PublicMode) == "1" {
		resText = "public mode enabled"
	}
	SendTextResponse(ctx, resText, fasthttp.StatusOK)
}

