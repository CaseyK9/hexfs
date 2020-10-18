package main

import (
	"github.com/valyala/fasthttp"
)

const (
	IsStandardKey = 0
	IsMasterKey = 1
	NotAuthorized = 2
)

func (b *BaseHandler) IsAuthorized(ctx *fasthttp.RequestCtx) bool {
	if b.GetAuthorizationLevel(ctx.Request.Header.Peek("authorization")) != IsMasterKey {
		SendTextResponse(ctx, "Not authorized.", fasthttp.StatusUnauthorized)
		return false
	}
	return true
}

func (b *BaseHandler)GetAuthorizationLevel(test []byte) int {
	switch string(test) {
	case b.Config.Security.MasterKey:
		return IsMasterKey
	case b.Config.Security.StandardKey:
		return IsStandardKey
	default:
		return NotAuthorized
	}
}