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
	VERSION = "1.0.1"
)

var SizeOfUploadDir int64

func main() {
	envErr := godotenv.Load()
	if envErr != nil {
		panic("Cannot find a .env file in the project root.")
	}
	ValidateEnv()
	if _, err := os.Stat(os.Getenv(UploadDirPath)); err != nil {
		if os.IsNotExist(err) {
			panic("Directory " + os.Getenv(UploadDirPath) + " does not exist. Create it and try again.")
		}
	}
	s, e := DirSize(os.Getenv(UploadDirPath))
	if e != nil {
		panic(e)
	}
	SizeOfUploadDir = s
	router := httprouter.New()
	server := http.Server{
		Addr: ":" + os.Getenv(Port),
		ReadHeaderTimeout: time.Second * 5000,
		WriteTimeout: time.Second * 5000,
		Handler: router,
	}
	router.GET("/:name", ServeIndex)
	router.POST("/", ServeUpload)
	router.POST("/delete/:name", ServeDelete)
	fmt.Println("Ready to serve requests on port " + os.Getenv(Port))
	log.Fatal(server.ListenAndServe())
}
