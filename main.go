package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
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
	FixedPort = "3030"
)

var (
	filesUploaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hexfs_uploaded",
		Help: "Files uploaded since the process started.",
	}, []string{"container"})
	heapUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hexfs_heap_usage",
		Help: "Memory in use by this process in bytes",
	}, []string{"container"})
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

func serve(ctx context.Context, r *mux.Router) (err error) {
	srv := &http.Server{
		Addr:         ":" + FixedPort,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: limit(applyCORS(r)),
	}
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	log.Printf("⬡ Server started and bound to port " + FixedPort)

	<-ctx.Done()

	log.Printf("⬡ Server has received a signal to stop.")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server shutdown failed: %s", err)
	}

	log.Printf("⬡ Server shut down gracefully.")

	if err == http.ErrServerClosed {
		err = nil
	}
	return
}

func main() {
	fmt.Println("hexFS file host software is now starting.")
	fmt.Print("You are running version " + VERSION + "\n\n")

	log.Println("⬡ Validating environment variables")
	ValidateEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	r := mux.NewRouter()

	log.Println("⬡ Configuring routes.")

	// Protected Routes (cannot be accessed without the key)
	r.HandleFunc("/file/delete/id/{id}", b.ProtectedRoute(b.ServeDelete)).Methods(http.MethodPost)
	r.HandleFunc("/file/delete/ip/{ip}", b.ProtectedRoute(b.ServeDelete)).Methods(http.MethodPost)
	r.HandleFunc("/file/delete/sha256/{sha256}", b.ProtectedRoute(b.ServeDelete)).Methods(http.MethodPost)

	// Conditional Routes (can be accessed without the key, but behavior is dynamic)
	r.HandleFunc("/file/info/{id}", b.ServeInformation).Methods(http.MethodGet)
	r.HandleFunc("/", b.ServeUpload).Methods(http.MethodPost, http.MethodOptions)

	// Public Routes (accessible without the key)
	r.HandleFunc("/auth/check", ServeCheckAuth).Methods(http.MethodGet)
	r.HandleFunc("/server/ping", ServePing).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", ServeFavicon).Methods(http.MethodGet)
	r.HandleFunc("/{id}", b.ServeFile).Methods(http.MethodGet)
	r.HandleFunc("/", ServeNotFound).Methods(http.MethodGet)
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { SendTextResponse(&w, "Method not allowed.", http.StatusMethodNotAllowed) })

	registry := prometheus.NewRegistry()
	registry.MustRegister(filesUploaded)
	registry.MustRegister(heapUsage)

	StartMetricsServer(registry)
	log.Println("⬡ Metrics server started and bound to port 3031.")

	InitRatelimiter()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		oscall := <-channel
		log.Printf("Received system call: %+v", oscall)
		cancel()
	}()
	if err := serve(ctx, r); err != nil {
		log.Printf("Failed to serve: +%v\n", err)
	}
}
