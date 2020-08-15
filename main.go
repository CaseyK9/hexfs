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
	VERSION = "1.1.0"
)

var SizeOfUploadDir int64

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv(Frontend) != "" {
		http.Redirect(w, r, os.Getenv(Frontend), 301)
	} else {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Page not found.",
		})
	}
}

func main() {
	envErr := godotenv.Load()
	if envErr != nil {
		panic("Cannot find a .env file in the project root.")
	}
	ValidateEnv()
	err := os.Mkdir(os.Getenv(UploadDirPath), 0755)
	if err != nil {
		if !os.IsExist(err) {
			panic("Directory " + os.Getenv(UploadDirPath) + " was attempted to be created by PSE, but failed. " + err.Error())
		}
		// is os.Exist is true then the directory already exists.
	}
	s, e := DirSize(os.Getenv(UploadDirPath))
	if e != nil {
		panic(e)
	}
	SizeOfUploadDir = s
	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Method not allowed.",
		})
	})
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NotFoundHandler(w, r)
	})

	router.GET("/", ServeIndex)
	router.GET("/:id", ServeIndex)
	router.GET("/:id/:name", ServeFile)
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
			router.ServeHTTP(w, r)
		}),
	}
	fmt.Println("Ready to serve requests on port " + os.Getenv(Port))
	log.Fatal(server.ListenAndServe())
}
