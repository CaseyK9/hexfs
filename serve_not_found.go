package main

import (
	"github.com/valyala/fasthttp"
)

func (b *BaseHandler)ServeNotFound(ctx *fasthttp.RequestCtx) {
	if ctx.Path()[0] == '/' && ctx.IsGet() && len(b.Config.Server.Frontend) != 0 {
		ctx.Redirect(b.Config.Server.Frontend, fasthttp.StatusTemporaryRedirect)
		return
	}
	SendTextResponse(ctx, "Page not found", fasthttp.StatusNotFound)
}
