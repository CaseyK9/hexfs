package main

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/valyala/fasthttp"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// DeleteFiles will delete files from both GCS and the database based on a filter of FileData.
func (b *BaseHandler) DeleteFiles(filter *FileData) (int64, error) {
	ext := path.Ext(filter.ID)
	if len(ext) != 0 {
		filter.ID = strings.TrimSuffix(filter.ID, ext)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := b.Database.Collection(MongoCollectionFiles).Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result FileData
		err := cur.Decode(&result)
		if err != nil {
			return 0, err

		}
		e := b.GCSClient.Bucket(os.Getenv(GCSBucketName)).Object(result.ID + result.Ext).Delete(ctx)
		if e != nil {
			// If object is not found, who cares, move on (might have been manually deleted from GCS?)
			if e != storage.ErrObjectNotExist {
				return 0, err
			}
		}
	}
	if err := cur.Err(); err != nil {
		return 0, err
	}

	rs, e := b.Database.Collection(MongoCollectionFiles).DeleteMany(ctx, filter)
	if e != nil {
		return 0, err
	}
	return rs.DeletedCount, nil
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
	deleted, err := b.DeleteFiles(q)
	if err != nil {
		SendTextResponse(ctx, "Error in deleting files: " + err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	SendTextResponse(ctx, strconv.FormatInt(deleted, 10), fasthttp.StatusOK)
	return
}