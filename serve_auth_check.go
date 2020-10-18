package main

import (
	"github.com/valyala/fasthttp"
)

// ServeCheckAuth validates either the standard or master key.
func (b *BaseHandler)ServeCheckAuth(ctx *fasthttp.RequestCtx) {
	if b.GetAuthorizationLevel(ctx.Request.Header.Peek("Authorization")) == NotAuthorized {
		SendTextResponse(ctx, "Not authorized.", fasthttp.StatusUnauthorized)
		return
	}
	SendNothing(ctx)
}


