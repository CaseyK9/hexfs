package main

import (
	"github.com/go-redis/redis_rate/v9"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/httputils/net"
	"time"
)
func (b *BaseHandler)limit(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		limiter := redis_rate.NewLimiter(b.RedisClient)
		res, err := limiter.Allow(ctx, net.GetIP(ctx), redis_rate.PerSecond(b.Config.Security.Ratelimit))
		if err != nil {
			SendTextResponse(ctx, "Failed to set rate limit: " + err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		if res.Allowed <= 0 {
			SendTextResponse(ctx, "You are being rate limited.", fasthttp.StatusTooManyRequests)
			return
		}
		h(ctx)
	}
}
func handleCORS(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Authorization")
		if ctx.Request.Header.IsOptions() {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		} else {
			h(ctx)
		}
	}
}

func (b *BaseHandler) handleHTTPRequest(ctx *fasthttp.RequestCtx) {

	switch string(ctx.Path()) {
	case "/upload":
		fasthttp.TimeoutHandler(b.ServeUpload, time.Minute * 15, "Upload timed out")(ctx)
		break
	case "/favicon.ico":
		ServeFavicon(ctx)
		break
	case "/auth/check":
		b.ServeCheckAuth(ctx)
		break
	case "/server/ping":
		b.ServePing(ctx)
		break
	default:
		if !ctx.IsGet() {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
		fasthttp.TimeoutHandler(b.ServeFile, time.Minute * 15, "Fetching file timed out")(ctx)
	}

}