package main

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"google.golang.org/api/option"
	"log"
	"time"
)

const (
	Version = "1.11.1"
	GCSKeyLoc = "./conf/key.json"
)

func main() {
	log.Print("hexFS " + Version + "\n\n")

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
	viper.SetDefault("server.idlen", 5)
	viper.SetDefault("server.concurrency", 128 * 4)
	viper.SetDefault("server.maxconnsperip", 16)
	viper.SetDefault("security.maxsizebytes", 52428800)
	viper.SetDefault("security.publicmode", false)
	viper.SetDefault("security.ratelimit", 2)

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(configuration.Security.MasterKey) == 0 {
		log.Fatal("STOP! At the very minimum, you must set master key in the environment. These were not set, so the program terminated for your safety.")
	}
	if configuration.Security.PublicMode {
		log.Println("Public mode is ENABLED. Authentication will not be required to upload!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     configuration.Net.Redis.URI,
		Password: configuration.Net.Redis.Password,
		DB:       configuration.Net.Redis.Db, 
	})

	status := redisClient.Ping(ctx).Err()
	if status != nil {
		log.Fatal("Could not ping Redis database: " + status.Error())
	}
	log.Println("Redis connection established")

	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(GCSKeyLoc))
	if err != nil {
		log.Fatal("Could not instantiate storage client: " + err.Error())
	}
	log.Println("Google Cloud Storage connection established")
	b := NewBaseHandler(c, redisClient, configuration)

	s := &fasthttp.Server{
		ErrorHandler:                  HandleError,
		Handler:                       b.limit(handleCORS(b.handleHTTPRequest)),
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Concurrency:                   configuration.Server.Concurrency,
		DisableKeepalive:              false,
		ReadTimeout:                   30 * time.Minute,
		WriteTimeout:                  30 * time.Minute,
		MaxConnsPerIP:                 configuration.Server.MaxConnsPerIP,
		TCPKeepalive:                  false,
		TCPKeepalivePeriod:            0,
		MaxRequestBodySize:            configuration.Security.MaxSizeBytes + (1024 * 1024),
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

	log.Println("-> Listening for new requests on port " + b.Config.Server.Port)
	if err = s.ListenAndServe(":" + b.Config.Server.Port); err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}

}
