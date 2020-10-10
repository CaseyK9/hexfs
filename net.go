package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
)

// GetIP get the IP from the header.
// It will try to fetch the client IP forwarded from Cloudflare.
func GetIP(ctx *fasthttp.RequestCtx) string {
	forwardedIP := ctx.Request.Header.Peek("X-Forwarded-For")
	if len(forwardedIP) != 0 {
		return string(forwardedIP)
	}
	forwardedIP = ctx.Request.Header.Peek("X-Real-IP")
	if len(forwardedIP) != 0 {
		return string(forwardedIP)
	}
	return ctx.RemoteIP().String()
}

func GetRoot(ctx *fasthttp.RequestCtx) string {
	protocol := "https"
	if strings.Contains(string(ctx.Request.Host()), "localhost:") {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s", protocol, ctx.Request.Host())
}