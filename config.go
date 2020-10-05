package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	MasterKey = "HFS_MASTER_KEY"
	Frontend = "HFS_FRONTEND"
	StandardKey = "HFS_STANDARD_KEY"
	PublicMode = "HFS_PUBLIC_MODE"
	MaxSizeBytes = "HFS_MAX_SIZE_BYTES"
	Endpoint = "HFS_ENDPOINT"
	GCSBucketName = "GCS_BUCKET_NAME"
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	GCSSecretKey = "GCS_SECRET_KEY"
	DisableFileBlacklist = "HFS_DISABLE_FILE_BLACKLIST"
	MongoConnectionURI = "HFS_MONGO_CONNECTION_URI"
	MongoDatabase = "HFS_MONGO_DATABASE"
	ContainerNickname = "HFS_CONTAINER_NICKNAME"
	FaviconLocation = "HFS_FAVICON_LOCATION"
)

// ValidateEnv validates the environment variables and throws log.Fatal if a variable is not correctly set or not set at all.
func ValidateEnv() {
	if os.Getenv(Endpoint) == os.Getenv(Frontend) {
		log.Fatal("Endpoint and frontend URLs cannot be the same because an infinite redirection loop would happen.")
	}
	for _, v := range []string{
		MasterKey,
		StandardKey,
		MaxSizeBytes,
		Endpoint,
		GCSBucketName,
		GoogleApplicationCredentials,
		GCSSecretKey,
		MongoConnectionURI,
		MongoDatabase,
		ContainerNickname,
	} {
		if len(os.Getenv(v)) == 0 {
			switch v {
			case MasterKey:
			case StandardKey:
			case Endpoint:
			case GCSBucketName:
			case GCSSecretKey:
			case GoogleApplicationCredentials:
			case MongoConnectionURI:
			case MongoDatabase:
				missing(v)
				break
			case MaxSizeBytes:
				log.Println("⬡ Setting max size of files to 50 MiB because it was not set")
				e := os.Setenv(MaxSizeBytes, "52428800")
				if e != nil { cannotSet(v) }
				break
			case ContainerNickname:
				n := "hexfs_" + strconv.FormatInt(time.Now().Unix(), 10)
				log.Println("⬡ Setting default container nickname to " + n + " because it was not set")
				e := os.Setenv(ContainerNickname, n)
				if e != nil { cannotSet(v) }
				break
			}
		}
	}
}

func missing(v string) {
	log.Fatal(v + " must be set in the environment variables")
}

func cannotSet(v string) {
	log.Fatal(v + " could not be set in the environment variables")
}