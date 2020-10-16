package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/hexfs/hlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/option"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	Version = "1.8.0"
	MongoCollectionFiles = "files"
	Gibibyte = 1073741824
	RedisKeyMaxCapacity = "maxcapacity"
	RedisKeyCurrentCapacity = "currentcapacity"
)

func main() {
	fmt.Println("Welcome to hexfs")
	fmt.Print("You are running version " + Version + "\n\n")

	//////////////////////////////////
	// Validate env
	///////////////////////////////////
	hlog.Log("env", hlog.LevelInfo, "Validating environment variables")
	ValidateEnv()

	//////////////////////////////////
	// Setup context
	///////////////////////////////////
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	//////////////////////////////////
	// Connect to MongoDB
	///////////////////////////////////
	hlog.Log("mongodb", hlog.LevelInfo, "Establishing MongoDB connection to database \"" + os.Getenv(MongoDatabase) + "\"")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv(MongoConnectionURI)))
	if err != nil {
		log.Fatal("Could not instantiate MongoDB client: " + err.Error())
	}
	hlog.Log("mongodb", hlog.LevelSuccess, "Connection established. Pinging database \"" + os.Getenv(MongoDatabase) + "\"")
	e := mongoClient.Ping(ctx, readpref.Primary())
	if e != nil {
		log.Fatal("Could not ping MongoDB database: " + e.Error())
	}
	hlog.Log("mongodb", hlog.LevelSuccess, "Ping successful.")

	//////////////////////////////////
	// Instantiate Redis
	///////////////////////////////////
	hlog.Log("redis", hlog.LevelInfo, "Establishing Redis connection.")
	dbInt := 0
	if len(os.Getenv(RedisDbInt)) > 0 {
		parsedInt, err := strconv.ParseInt(os.Getenv(RedisDbInt), 10, 64)
		if err != nil {
			log.Fatal("⬡ Failed to parse Redis db integer: " + err.Error())
		}
		dbInt = int(parsedInt)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv(RedisConnectionURI),
		Password: os.Getenv(RedisPassword), // no password set
		DB:       dbInt,  // use default DB
	})

	//////////////////////////////////
	// Ping Redis
	///////////////////////////////////
	hlog.Log("redis", hlog.LevelInfo, "Pinging Redis database.")
	status := redisClient.Ping(ctx).Err()
	if status != nil {
		log.Fatal("⬡ Could not ping Redis database: " + status.Error())
	}
	hlog.Log("redis", hlog.LevelSuccess, "Ping successful.")

	//////////////////////////////////
	// Check if max capacity size already set
	///////////////////////////////////
	res, err := redisClient.Get(ctx, RedisKeyMaxCapacity).Result()
	if err == redis.Nil {
		maxCapacity, err := strconv.ParseInt(os.Getenv(MaxCapacity), 10, 64)
		if err != nil {
			log.Fatal("⬡ Failed to parse max capacity integer: " + err.Error())
		}
		hlog.Log("capacity", hlog.LevelInfo, fmt.Sprintf("Max capacity will be set to %d GiB because it was not set", maxCapacity))

		calcMaxCapacity := maxCapacity * Gibibyte
		err = redisClient.Set(ctx, RedisKeyMaxCapacity, calcMaxCapacity, 0).Err()
		if err != nil {
			log.Fatal("⬡ Failed to set max capacity in Redis database: " + err.Error())
		}
	} else if err != nil {
		log.Fatal("⬡ Failed to get max capacity in Redis database: " + err.Error())
	} else {
		hlog.Log("capacity", hlog.LevelInfo, fmt.Sprintf("Max capacity already set to %s GiB, not changing", res))
	}

	//////////////////////////////////
	// Check if current capacity size already set
	///////////////////////////////////
	res, err = redisClient.Get(ctx, RedisKeyCurrentCapacity).Result()
	if err == redis.Nil {
		hlog.Log("capacity", hlog.LevelInfo, "Getting current capacity because it was not set")
		hlog.Log("capacity", hlog.LevelInfo, "Aggregating current capacity from data base (this may take a while...)")
		aggCtx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
		defer cancel()

		op1 := bson.D{
			{"$group", bson.D{
				{"_id", nil},
				{"total", bson.D{
					{"$sum", "$size"},
				}},
			}},
		}
		opList := []bson.D{op1}
		cursor, err := mongoClient.Database(os.Getenv(MongoDatabase)).Collection(MongoCollectionFiles).Aggregate(aggCtx, opList)
		if err != nil {
			log.Fatal("⬡ Failed to iterate: " + err.Error())
		}
		defer cursor.Close(aggCtx)
		var results []bson.M
		if err = cursor.All(aggCtx, &results); err != nil {
			log.Fatal(err)
		}
		for _, result := range results {
			err = redisClient.Set(ctx, RedisKeyCurrentCapacity, result["total"], 0).Err()
			if err != nil {
				log.Fatal("⬡ Failed to set current capacity in Redis database: " + err.Error())
			}
			hlog.Log("capacity", hlog.LevelSuccess, "Finished aggregation.")
			break
		}
	} else if err != nil {
		log.Fatal("⬡ Failed to get current capacity in Redis database: " + err.Error())
	} else {
		hlog.Log("capacity", hlog.LevelInfo, fmt.Sprintf("Current capacity already set to %s bytes, not changing", res))
	}

	//////////////////////////////////
	// Connect to Google Cloud Storage
	///////////////////////////////////
	hlog.Log("gcs", hlog.LevelInfo, "Establishing Google Cloud Storage client with key file")
	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(os.Getenv(GoogleApplicationCredentials)))
	if err != nil {
		log.Fatal("Could not instantiate storage client: " + err.Error())
	}
	defer c.Close()
	b := NewBaseHandler(mongoClient.Database(os.Getenv(MongoDatabase)), c, redisClient)

	//////////////////////////////////
	// Start HTTP server
	///////////////////////////////////
	s := &fasthttp.Server{
		Handler:                            b.limit(handleCORS(b.handleHTTPRequest)),
		ErrorHandler:                       nil,
		HeaderReceived:                     nil,
		ContinueHandler:                    nil,
		Name:                               "hexfs v" + Version,
		Concurrency:                        128 * 4,
		DisableKeepalive:                   false,
		ReadTimeout:                        20 * time.Minute,
		WriteTimeout:                       20 * time.Minute,
		MaxConnsPerIP:                      256,
		TCPKeepalive:                       false,
		TCPKeepalivePeriod:                 0,
		MaxRequestBodySize:                 int(b.MaxSizeBytes) + 1024,
		ReduceMemoryUsage:                  false,
		GetOnly:                            false,
		DisablePreParseMultipartForm:       false,
		LogAllErrors:                       false,
		DisableHeaderNamesNormalizing:      false,
		NoDefaultServerHeader:              false,
		NoDefaultDate:                      false,
		NoDefaultContentType:               false,
		KeepHijackedConns:                  false,
	}

	p := os.Getenv(Port)
	if len(p) == 0 {
		p = "3030"
	}

	hlog.Log("all done", hlog.LevelSuccess, "Start-up complete, server is ready to accept requests at port " + p)
	if err = s.ListenAndServe(":" + os.Getenv(Port)); err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}
}
