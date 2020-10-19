package main

import (
	"cloud.google.com/go/storage"
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"
	"path"
	"strconv"
	"strings"
)

// DeleteFiles will delete files from both GCS and the database based on a filter of FileData.
func (b *BaseHandler) DeleteFiles(ctx *fasthttp.RequestCtx, filter *FileData) (int64, int64, error) {
	ext := path.Ext(filter.ID)
	if len(ext) != 0 {
		filter.ID = strings.TrimSuffix(filter.ID, ext)
	}
	cur, err := b.Database.Collection(MongoCollectionFiles).Find(ctx, filter)
	if err != nil {
		return 0, 0, err
	}
	defer cur.Close(ctx)

	sizeDeleted := int64(0)
	for cur.Next(ctx) {
		var result FileData
		err := cur.Decode(&result)
		if err != nil {
			return 0, 0, err
		}
		e := b.GCSClient.Bucket(b.Config.Net.GCS.BucketName).Object(result.ID + result.Ext).Delete(ctx)
		if e != nil {
			// If object is not found, who cares, move on (might have been manually deleted from GCS?)
			if e != storage.ErrObjectNotExist {
				return 0, 0, err
			}
		}
		sizeDeleted += result.Size
	}
	if err := cur.Err(); err != nil {
		return 0, 0, err
	}

	rs, e := b.Database.Collection(MongoCollectionFiles).DeleteMany(ctx, filter)
	if e != nil {
		return 0, 0, err
	}
	return rs.DeletedCount, sizeDeleted, nil
}

func (b *BaseHandler) ServeDelete(ctx *fasthttp.RequestCtx) {
	q := &FileData{
		ID:                string(ctx.QueryArgs().Peek("id")),
		SHA256:            string(ctx.QueryArgs().Peek("sha256")),
		IP:                string(ctx.QueryArgs().Peek("ip")),
	}
	if len(q.ID) == 0 && len(q.SHA256) == 0 && len(q.IP) == 0 {
		SendTextResponse(ctx, "Nothing provided to delete.", fasthttp.StatusBadRequest)
		return
	}
	deleted, bytesDeleted, err := b.DeleteFiles(ctx, q)
	if err != nil {
		SendTextResponse(ctx, "Error in deleting files: " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	r, e := b.RedisClient.Get(ctx, RedisKeyCurrentCapacity).Result()
	if e == redis.Nil {
		SendTextResponse(ctx, "Current capacity is unknown, cannot proceed.", fasthttp.StatusInternalServerError)
		return
	} else if e != nil {
		SendTextResponse(ctx, "Failed to get current capacity. " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}
	parsed, e := strconv.ParseInt(r, 10, 64)
	if e != nil {
		SendTextResponse(ctx, "Failed to parse current capacity. " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}

	e = b.RedisClient.Set(ctx, RedisKeyCurrentCapacity, parsed - bytesDeleted, 0).Err()
	if e != nil {
		SendTextResponse(ctx, "Failed to update current capacity. " + e.Error(), fasthttp.StatusInternalServerError)
		return
	}

	SendTextResponse(ctx, strconv.FormatInt(deleted, 10), fasthttp.StatusOK)
	return
}