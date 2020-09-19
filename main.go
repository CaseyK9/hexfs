package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	VERSION = "v1.5.0"
)

func main() {
	fmt.Println("=======\nhexFS " + VERSION + "\n=======")
	fmt.Println("Checking for .env file")
	envErr := godotenv.Load()
	if envErr != nil {
		panic("Cannot find a .env file in the project root.")
	}
	fmt.Println("Validating environment variables")
	ValidateEnv()
	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SendTextResponse(&w, "Method not allowed.", http.StatusMethodNotAllowed)
	})
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeNotFound(w, r)
	})
	router.GET("/", ServeIndex)
	router.GET("/:id", ServeIndex)
	router.POST("/", ServeUpload)
	router.POST("/delete/:name", ServeDelete)
	server := http.Server{
		Addr: ":" + os.Getenv(Port),
		ReadHeaderTimeout: time.Second * 5000,
		WriteTimeout: time.Second * 5000,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.Method == "POST" {
				if os.Getenv(UploadKey) != r.Header.Get("authorization") {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
			router.ServeHTTP(w, r)
		}),
		}
	fmt.Println("All done! Serving requests on port " + os.Getenv(Port))
	log.Fatal(server.ListenAndServe())
}
