package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

const (
	Version = "1.7.0"
	MongoCollectionFiles = "files"
	FixedPort = "3030"
)

func handleCORS(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Authorization")
		if ctx.Request.Header.IsOptions() {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		} else {
			h(ctx)
		}
	}
}

func (b *BaseHandler) handleHTTPRequest(ctx *fasthttp.RequestCtx) {

	switch string(ctx.Path()) {
	case "/upload":
		fasthttp.TimeoutHandler(b.ServeUpload, time.Minute * 15, "Upload timed out")(ctx)
		break
	case "/favicon.ico":
		ServeFavicon(ctx)
		break
	case "/file/delete":
		if !b.IsAuthorized(ctx) {
			return
		}
		fasthttp.TimeoutHandler(b.ServeDelete, time.Minute * 5, "Deleting files timed out")(ctx)
		break
	case "/file/info":
		fasthttp.TimeoutHandler(b.ServeInformation, time.Second * 15, "File into retrieval timed out")(ctx)
		break
	case "/auth/check":
		ServeCheckAuth(ctx)
		break
	case "/server/ping":
		ServePing(ctx)
		break
	default:
		if !ctx.IsGet() {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
		fasthttp.TimeoutHandler(b.ServeFile, time.Minute * 3, "Fetching file timed out")(ctx)
	}

}
func main() {
	fmt.Println("hexfs file host software is now starting.")
	fmt.Print("You are running version " + Version + "\n\n")

	log.Println("⬡ Validating environment variables")
	ValidateEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	log.Println("⬡ Establishing MongoDB connection to database \"" + os.Getenv(MongoDatabase) + "\"")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv(MongoConnectionURI)))
	if err != nil {
		log.Fatal("Could not instantiate MongoDB client: " + err.Error())
	}
	log.Println("⬡ Connection established. Pinging database \"" + os.Getenv(MongoDatabase) + "\"")
	e := mongoClient.Ping(ctx, readpref.Primary())
	if e != nil {
		log.Fatal("Could not ping MongoDB database: " + e.Error())
	}
	log.Println("⬡ Ping successful.")

	log.Println("⬡ Establishing Google Cloud Storage client with key file")
	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(os.Getenv(GoogleApplicationCredentials)))
	if err != nil {
		log.Fatal("Could not instantiate storage client: " + err.Error())
	}
	defer c.Close()
	b := NewBaseHandler(mongoClient.Database(os.Getenv(MongoDatabase)), c)

	s := &fasthttp.Server{
		Handler:                            handleCORS(b.handleHTTPRequest),
		ErrorHandler:                       nil,
		HeaderReceived:                     nil,
		ContinueHandler:                    nil,
		Name:                               "hexfs v" + Version,
		Concurrency:                        128 * 256,
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
	if err = s.ListenAndServe(":" + FixedPort); err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}
}
