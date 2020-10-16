package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"
)

type capacityStats struct {
	CurrentCapacity string `json:"current"`
	MaxCapacity string `json:"max"`
}

func (b *BaseHandler) ServeCapacity(ctx *fasthttp.RequestCtx) {
	currentCap, err := b.RedisClient.Get(ctx, RedisKeyCurrentCapacity).Result()
	if err == redis.Nil {
		SendTextResponse(ctx, "Current capacity unknown", fasthttp.StatusInternalServerError)
		return
	} else if err != nil {
		SendTextResponse(ctx, "Failed to determine the current capacity of the host. " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	maxCap, err := b.RedisClient.Get(ctx, RedisKeyMaxCapacity).Result()
	if err == redis.Nil {
		SendTextResponse(ctx, "Maximum capacity unknown", fasthttp.StatusInternalServerError)
		return
	} else if err != nil {
		SendTextResponse(ctx, "Failed to determine the maximum capacity of the host. " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	SendJSONResponse(ctx, capacityStats{
		CurrentCapacity: currentCap,
		MaxCapacity: maxCap,
	})
}


