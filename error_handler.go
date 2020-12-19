package main

import (
	"github.com/valyala/fasthttp"
)

func HandleError(ctx *fasthttp.RequestCtx, err error) {
	switch err {
	default:
		SendTextResponse(ctx, "Sorry! An internal server error has occurred.", fasthttp.StatusInternalServerError)
		break
	case fasthttp.ErrBodyTooLarge:
		SendTextResponse(ctx, "Upload size too large.", fasthttp.StatusBadRequest)
		break
	case fasthttp.ErrNoFreeConns:
		SendTextResponse(ctx, "There are currently too many requests from this host.", fasthttp.StatusTooManyRequests)
		break
	case fasthttp.ErrPerIPConnLimit:
		SendTextResponse(ctx, "There are currently too many requests from this IP.", fasthttp.StatusTooManyRequests)
		break
	case fasthttp.ErrConcurrencyLimit:
		SendTextResponse(ctx, "The server is temporarily overloaded. Try again later.", fasthttp.StatusTooManyRequests)
		break
	}
}