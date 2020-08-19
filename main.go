package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	VERSION = "v1.3.1"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv(Frontend) != "" {
		http.Redirect(w, r, os.Getenv(Frontend), http.StatusPermanentRedirect)
	} else {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Page not found.",
		}, http.StatusNotFound)
	}
}

func CheckForUpdates() {
	resp, err := http.Get("https://api.github.com/repos/ethanwritescode/hexfs/releases/latest")
	if err != nil {
		fmt.Println("Warning: Could not check for updates from GitHub")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Could not read Github response body.")
		return
	}

	var r GithubRelease
	e := json.Unmarshal(body, &r)
	if e != nil {
		fmt.Println("Could not parse Github response body to JSON.")
		return
	}
	if r.TagName != VERSION {
		_, _ = fmt.Fprintf(os.Stdout, "\n\n========\nUpdate available! (%s) -> (%s)\nDownload: %s\n=======\n\n", VERSION, r.TagName, r.HTMLURL)
	} else {
		fmt.Println("You have the most up-to-date version. (" + VERSION + ")")
	}
}

func main() {
	fmt.Println("=======\nhexFS v" + VERSION + "\n=======")
	fmt.Println("Checking for updates")
	CheckForUpdates()
	fmt.Println("Checking for .env file")
	envErr := godotenv.Load()
	if envErr != nil {
		panic("Cannot find a .env file in the project root.")
	}
	fmt.Println("Validating environment variables")
	ValidateEnv()
	fmt.Println("Making directory " + os.Getenv(UploadDirPath) + " if it doesn't exist")
	err := os.Mkdir(os.Getenv(UploadDirPath), 0755)
	if err != nil {
		if !os.IsExist(err) {
			panic("Directory " + os.Getenv(UploadDirPath) + " was going to be created by hexFS, but failed. " + err.Error())
		}
		// is os.Exist is true then the directory already exists.
	}
	fmt.Println("Getting initial size of upload directory path")
	//lmt := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute * 30})
	//lmt.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	//lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
	//	_ = json.NewEncoder(w).Encode(ResponseError{
	//		Status:  1,
	//		Message: "You are being rate limited.",
	//	})
	//})
	//lmt.SetMessage("")
	//lmt.SetMessageContentType("application/json")

	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Method not allowed.",
		}, http.StatusMethodNotAllowed)
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
		//Handler: tollbooth.LimitFuncHandler(lmt, func(w http.ResponseWriter, r *http.Request) {
		//	w.Header().Set("Access-Control-Allow-Origin", "*")
		//	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		//	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		//	router.ServeHTTP(w, r)
		//}),
		Handler: limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			router.ServeHTTP(w, r)
		})),
	}
	fmt.Println("Initializing ratelimiter cleanup goroutine")
	InitRatelimiter()
	fmt.Println("Serving requests on port " + os.Getenv(Port))
	log.Fatal(server.ListenAndServe())
}
