package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// GetFileData retrieves file data based on a query in the format of the FileData struct. It will return *FileData if an object is found or an error if it's not.
func (b *BaseHandler) GetFileData(query FileData) (*FileData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()

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