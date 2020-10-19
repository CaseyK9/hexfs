package main

import (
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetFileData retrieves file data based on a query in the format of the FileData struct. It will return *FileData if an object is found or an error if it's not.
func (b *BaseHandler) GetFileData(ctx *fasthttp.RequestCtx, query FileData) (*FileData, error) {
	fd := &FileData{}
	e := b.Database.Collection(MongoCollectionFiles).FindOne(ctx, query).Decode(&fd)
	if e != nil {
		if e == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, e
	}
	return fd, nil
}