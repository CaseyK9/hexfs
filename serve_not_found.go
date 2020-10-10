package main

import (
	"github.com/valyala/fasthttp"
	"os"
)

func ServeNotFound(ctx *fasthttp.RequestCtx) {
	if ctx.Path()[0] == '/' && ctx.IsGet() && len(os.Getenv(Frontend)) != 0 {
		ctx.Redirect(os.Getenv(Frontend), fasthttp.StatusTemporaryRedirect)
		return
	}
	SendTextResponse(ctx, "Page not found", fasthttp.StatusNotFound)
}
