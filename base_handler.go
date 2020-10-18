package main

import (
	"cloud.google.com/go/storage"
	"encoding/base64"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type BaseHandler struct {
	Database *mongo.Database
	GCSClient *storage.Client
	RedisClient *redis.Client
	Key []byte
	Config Configuration
}

func NewBaseHandler(db *mongo.Database, gcsClient *storage.Client, redisClient *redis.Client, c Configuration) *BaseHandler {
	k, e := base64.StdEncoding.DecodeString(c.Net.GCS.SecretKey)
	if e != nil {
		log.Fatal("Key not properly formatted to Base64.")
	}

	return &BaseHandler{
		Database: db,
		GCSClient: gcsClient,
		Key: k,
		RedisClient: redisClient,
		Config: c,
	}
}