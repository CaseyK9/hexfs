package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
)

// GetIP get the IP from the header
func GetIP(ctx *fasthttp.RequestCtx) string {
	return string(ctx.RemoteIP())
}

func GetRoot(ctx *fasthttp.RequestCtx) string {
	protocol := "https"
	if strings.Contains(string(ctx.Request.Host()), "localhost:") {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s", protocol, ctx.Request.Host())
}