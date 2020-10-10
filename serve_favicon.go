package main

import (
	"github.com/valyala/fasthttp"
	"io"
	"os"
)

func ServeFavicon(ctx *fasthttp.RequestCtx) {
	if len(os.Getenv(FaviconLocation)) == 0 {
		ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	f, e := os.OpenFile(os.Getenv(FaviconLocation), os.O_RDONLY, 0666)
	if e != nil {
		if e == os.ErrNotExist {
			ctx.Response.SetStatusCode(fasthttp.StatusNoContent)
		} else {
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		}
		return
	}
	defer f.Close()
	_, e = io.Copy(ctx.Response.BodyWriter(), f)
	if e != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}
