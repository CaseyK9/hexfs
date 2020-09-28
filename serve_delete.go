package main

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strconv"
	"time"
)

// DeleteFiles will delete files from both GCS and the database based on a filter of FileData.
func (b *BaseHandler) DeleteFiles(filter FileData) (int64, error) {
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
			// If object is not found, who cares, move on
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

func (b *BaseHandler) ServeDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	deleted, err := b.DeleteFiles(FileData{
		ID:                vars["id"],
		SHA256:            vars["sha256"],
		IP:                vars["ip"],
	})
	if err != nil {
		SendTextResponse(&w, "Error in deleting files: " + err.Error(), http.StatusInternalServerError)
		return
	}

	SendTextResponse(&w, strconv.FormatInt(deleted, 10), http.StatusOK)
	return
}