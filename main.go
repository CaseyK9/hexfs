package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/vysiondev/hexfs/hlog"
	"google.golang.org/api/option"
	"log"
	"time"
)

const (
	Version = "1.10.1"
	GCSKeyLoc = "./conf/key.json"
)

func main() {
	fmt.Println("Welcome to hexfs")
	fmt.Print("You are running version " + Version + "\n\n")

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
	viper.SetDefault("server.idlen", 5)
	viper.SetDefault("security.maxsizebytes", 52428800)
	viper.SetDefault("security.publicmode", false)

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
	if len(configuration.Security.MasterKey) == 0 {
		log.Fatal("STOP! At the very minimum, you must set master key in the environment. These were not set, so the program terminated for your safety.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	//////////////////////////////////
	// Instantiate Redis
	///////////////////////////////////
	hlog.Log("redis", hlog.LevelInfo, "Establishing Redis connection.")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     configuration.Net.Redis.URI,
		Password: configuration.Net.Redis.Password,
		DB:       configuration.Net.Redis.Db, 
	})

	//////////////////////////////////
	// Ping Redis
	///////////////////////////////////
	hlog.Log("redis", hlog.LevelInfo, "Pinging Redis database.")
	status := redisClient.Ping(ctx).Err()
	if status != nil {
		log.Fatal("Could not ping Redis database: " + status.Error())
	}
	hlog.Log("redis", hlog.LevelSuccess, "Ping successful.")

	//////////////////////////////////
	// Connect to Google Cloud Storage
	///////////////////////////////////
	hlog.Log("gcs", hlog.LevelInfo, "Establishing Google Cloud Storage client with key file")
	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(GCSKeyLoc))
	if err != nil {
		log.Fatal("Could not instantiate storage client: " + err.Error())
	}
	b := NewBaseHandler(c, redisClient, configuration)

	//////////////////////////////////
	// Start HTTP server
	///////////////////////////////////
	s := &fasthttp.Server{
		Handler:                       b.limit(handleCORS(b.handleHTTPRequest)),
		ErrorHandler:                  nil,
		HeaderReceived:                nil,
		ContinueHandler:               nil,
		Name:                          "hexfs v" + Version,
		Concurrency:                   128 * 4,
		DisableKeepalive:              false,
		ReadTimeout:                   30 * time.Minute,
		WriteTimeout:                  30 * time.Minute,
		MaxConnsPerIP:                 16,
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
