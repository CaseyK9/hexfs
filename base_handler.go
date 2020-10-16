package main

import (
	"cloud.google.com/go/storage"
	"encoding/base64"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"strconv"
)

type BaseHandler struct {
	Database *mongo.Database
	GCSClient *storage.Client
	RedisClient *redis.Client
	Key []byte
	MaxSizeBytes int64
}

func NewBaseHandler(db *mongo.Database, gcsClient *storage.Client, redisClient *redis.Client) *BaseHandler {
	k, e := base64.StdEncoding.DecodeString(os.Getenv(GCSSecretKey))
	if e != nil {
		log.Fatal("Key not properly formatted to Base64.")
	}
	i, e := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	if e != nil {
		log.Fatal("Cannot parse the max size in bytes for a file.")
	}

	return &BaseHandler{
		Database: db,
		GCSClient: gcsClient,
		Key: k,
		MaxSizeBytes: i,
		RedisClient: redisClient,
	}
}