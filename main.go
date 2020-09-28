package main

import (
	"cloud.google.com/go/storage"
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	VERSION = "1.6.0"
	MongoCollectionFiles = "files"
)

func applyCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		if r.Method == http.MethodOptions {
			SendNothing(&w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("                     #######  #####  \n#    # ###### #    # #       #     # \n#    # #       #  #  #       #       \n###### #####    ##   #####    #####  \n#    # #        ##   #             # \n#    # #       #  #  #       #     # \n#    # ###### #    # #        #####  ")
	fmt.Println("\nYou are running version " + VERSION)
	fmt.Println("\n⬡ Checking for .env file")
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Cannot find a .env file in the project root. Create one, set the values specified in the README, and retry.")
	}
	fmt.Println("⬡ Validating environment variables")
	ValidateEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("⬡ Establishing MongoDB connection to database \"" + os.Getenv(MongoDatabase) + "\"")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv(MongoConnectionURI)))
	if err != nil {
		log.Fatal("Could not instantiate MongoDB client: " + err.Error())
	}
	fmt.Println("⬡ Connection established. Pinging database \"" + os.Getenv(MongoDatabase) + "\"")
	e := mongoClient.Ping(ctx, readpref.Primary())
	if e != nil {
		log.Fatal("Could not ping MongoDB database: " + e.Error())
	}
	fmt.Println("⬡ Ping successful.")

	fmt.Println("⬡ Establishing Google Cloud Storage client with key file")
	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(os.Getenv(GoogleApplicationCredentials)))
	if err != nil {
		log.Fatal("Could not instantiate storage client: " + err.Error())
	}
	defer c.Close()
	b := NewBaseHandler(mongoClient.Database(os.Getenv(MongoDatabase)), c)
	r := mux.NewRouter()

	fmt.Println("⬡ Configuring routes.")

	// Protected Routes (cannot be accessed without the key)
	r.HandleFunc("/file/delete/id/{id}", b.ProtectedRoute(b.ServeDelete)).Methods(http.MethodPost)
	r.HandleFunc("/file/delete/ip/{ip}", b.ProtectedRoute(b.ServeDelete)).Methods(http.MethodPost)
	r.HandleFunc("/file/delete/sha256/{sha256}", b.ProtectedRoute(b.ServeDelete)).Methods(http.MethodPost)
	r.HandleFunc("/", b.ProtectedRoute(b.ServeUpload)).Methods(http.MethodPost, http.MethodOptions)

	// Conditional Routes (can be accessed without the key, but limited information is returned)
	r.HandleFunc("/file/info/{id}", b.ServeInformation).Methods(http.MethodGet)

	// Public Routes (accessible without the key)
	r.HandleFunc("/auth/check", ServeCheckAuth).Methods(http.MethodGet)
	r.HandleFunc("/server/ping", ServePing).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", ServeFavicon).Methods(http.MethodGet)
	r.HandleFunc("/{id}", b.ServeFile).Methods(http.MethodGet)

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ServeNotFound(w, r) })
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { SendTextResponse(&w, "Method not allowed.", http.StatusMethodNotAllowed) })
	srv := &http.Server{
		Addr:         ":" + os.Getenv(Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: limit(applyCORS(r)),
	}
	InitRatelimiter()
	fmt.Println("⬡ Done! Bound to port " + os.Getenv(Port))

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second * 15, "Wait for all requests to finish")
	flag.Parse()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

	ctx, cancel = context.WithTimeout(context.Background(), wait)
	defer cancel()
	_ = srv.Shutdown(ctx)
	fmt.Println(" ▶ Shutting down.")
	os.Exit(0)
}
