package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/hexfs/hlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/option"
	"log"
	"time"
)

const (
	Version = "1.9.0"
	MongoCollectionFiles = "files"
	Gibibyte = 1073741824
	RedisKeyMaxCapacity = "maxcapacity"
	RedisKeyCurrentCapacity = "currentcapacity"
	GCSKeyLoc = "./conf/key.json"
)

func main() {
	fmt.Println("Welcome to hexfs")
	fmt.Print("You are running version " + Version + "\n\n")

	//////////////////////////////////
	// Setup env
	///////////////////////////////////
	hlog.Log("env", hlog.LevelInfo, "Setting up env")
	viper.SetConfigName("config")
	viper.AddConfigPath("./conf/")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	var configuration Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Set undefined variables
	viper.SetDefault("server.port", "3030")
	viper.SetDefault("net.redis.db", 0)
	viper.SetDefault("security.maxsizebytes", 52428800)
	viper.SetDefault("security.capacity", 5 * Gibibyte)
	viper.SetDefault("security.disablefileblacklist", false)
	viper.SetDefault("security.publicmode", false)

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(configuration.Security.StandardKey) == 0 || len(configuration.Security.MasterKey) == 0 {
		log.Fatal("STOP! At the very minimum, you must set the standard and master keys in the environment. These were not set, so the program terminated for your safety.")
	}
	//////////////////////////////////
	// Setup context
	///////////////////////////////////
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	//////////////////////////////////
	// Connect to MongoDB
	///////////////////////////////////
	hlog.Log("mongodb", hlog.LevelInfo, "Establishing MongoDB connection to database \"" + configuration.Net.Mongo.URI + "\"")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(configuration.Net.Mongo.URI))
	if err != nil {
		log.Fatal("Could not instantiate MongoDB client: " + err.Error())
	}
	hlog.Log("mongodb", hlog.LevelSuccess, "Connection established. Pinging instance")
	e := mongoClient.Ping(ctx, readpref.Primary())
	if e != nil {
		log.Fatal("Could not ping MongoDB database: " + e.Error())
	}
	hlog.Log("mongodb", hlog.LevelSuccess, "Ping successful.")

	//////////////////////////////////
	// Instantiate Redis
	///////////////////////////////////
	hlog.Log("redis", hlog.LevelInfo, "Establishing Redis connection.")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     configuration.Net.Redis.URI,
		Password: configuration.Net.Redis.Password, // no password set
		DB:       configuration.Net.Redis.Db,  // use default DB
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
		hlog.Log("capacity", hlog.LevelInfo, fmt.Sprintf("Max capacity will be set to %d GiB because it was not set", configuration.Security.Capacity))

		err = redisClient.Set(ctx, RedisKeyMaxCapacity, configuration.Security.Capacity * Gibibyte, 0).Err()
		if err != nil {
			log.Fatal("⬡ Failed to set max capacity in Redis database: " + err.Error())
		}
	} else if err != nil {
		log.Fatal("⬡ Failed to get max capacity in Redis database: " + err.Error())
	} else {
		hlog.Log("capacity", hlog.LevelInfo, fmt.Sprintf("Max capacity already set to %s bytes, not changing", res))
	}

	//////////////////////////////////
	// Check if current capacity size already set
	///////////////////////////////////
	res, err = redisClient.Get(ctx, RedisKeyCurrentCapacity).Result()
	if err == redis.Nil {
		hlog.Log("capacity", hlog.LevelInfo, "Setting current capacity because it was not set before")
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
		cursor, err := mongoClient.Database(configuration.Net.Mongo.Database).Collection(MongoCollectionFiles).Aggregate(aggCtx, opList)
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
	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(GCSKeyLoc))
	if err != nil {
		log.Fatal("Could not instantiate storage client: " + err.Error())
	}
	defer c.Close()
	b := NewBaseHandler(mongoClient.Database(configuration.Net.Mongo.Database), c, redisClient, configuration)

	//////////////////////////////////
	// Start HTTP server
	///////////////////////////////////
	s := &fasthttp.Server{
		Handler:                            b.limit(handleCORS(b.handleHTTPRequest)),
		ErrorHandler:                       nil,
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Name:                          "hexfs v" + Version,
		Concurrency:                   128 * 4,
		DisableKeepalive:              false,
		ReadTimeout:                   20 * time.Minute,
		WriteTimeout:                  20 * time.Minute,
		MaxConnsPerIP:                 256,
		TCPKeepalive:                  false,
		TCPKeepalivePeriod:            0,
		MaxRequestBodySize:            configuration.Security.MaxSizeBytes + 1024,
		ReduceMemoryUsage:             false,
		GetOnly:                       false,
		DisablePreParseMultipartForm:  false,
		LogAllErrors:                  false,
		DisableHeaderNamesNormalizing: false,
		NoDefaultServerHeader:         false,
		NoDefaultDate:                 false,
		NoDefaultContentType:          false,
		KeepHijackedConns:             false,
	}

	hlog.Log("all done", hlog.LevelSuccess, "Start-up complete, server is ready to accept requests at port " + b.Config.Server.Port)
	if err = s.ListenAndServe(":" + b.Config.Server.Port); err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}
}
