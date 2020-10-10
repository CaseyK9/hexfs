package main

import (
	"github.com/valyala/fasthttp"
	"os"
)

const (
	IsStandardKey = 0
	IsMasterKey = 1
	NotAuthorized = 2
)

func (b *BaseHandler) IsAuthorized(ctx *fasthttp.RequestCtx) bool {
	if GetAuthorizationLevel(ctx.Request.Header.Peek("authorization")) != IsMasterKey {
		SendTextResponse(ctx, "Not authorized.", fasthttp.StatusUnauthorized)
		return false
	}
	return true
}

func GetAuthorizationLevel(test []byte) int {
	switch string(test) {
	case os.Getenv(MasterKey):
		return IsMasterKey
	case os.Getenv(StandardKey):
		return IsStandardKey
	default:
		return NotAuthorized
	}
}