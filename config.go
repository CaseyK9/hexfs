package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	Port = "HFS_PORT"
	MasterKey = "HFS_MASTER_KEY"
	StandardKey = "HFS_STANDARD_KEY"
	PublicMode = "HFS_PUBLIC_MODE"
	MaxSizeBytes = "HFS_MAX_SIZE_BYTES"
	Endpoint = "HFS_ENDPOINT"
	Frontend = "HFS_FRONTEND"
	GCSBucketName = "GCS_BUCKET_NAME"
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	GCSSecretKey = "GCS_SECRET_KEY"
	DisableFileBlacklist = "HFS_DISABLE_FILE_BLACKLIST"
	MongoConnectionURI = "HFS_MONGO_CONNECTION_URI"
	MongoDatabase = "HFS_MONGO_DATABASE"
)

// ValidateEnv validates the environment variables and throws log.Fatal if a variable is not correctly set or not set at all.
func ValidateEnv() {
	for _, v := range []string{
		Port,
		StandardKey,
		MasterKey,
		MaxSizeBytes,
		Endpoint,
		PublicMode,
		GCSBucketName,
		GoogleApplicationCredentials,
		DisableFileBlacklist,
		MongoConnectionURI,
		MongoDatabase,
	} {
		switch v {
		case DisableFileBlacklist, PublicMode:
			if os.Getenv(v) == "" || os.Getenv(v) != "1" {
				e := os.Setenv(v, "0")
				if e != nil {
					log.Fatal("Default value of " + v + " could not be set to 0.")
				}
			}
			break
		case GCSBucketName, GoogleApplicationCredentials, GCSSecretKey:
			if os.Getenv(v) == "" {
				log.Fatal("You must set the proper Google Cloud Storage variables.")
			}
			break
		case Port:
			if os.Getenv(v) == "" {
				e := os.Setenv(v, "7250")
				if e != nil {
					log.Fatal("Could not set default port to 7250")
				}
			} else {
				n, e := strconv.ParseInt(os.Getenv(v), 0, 64)
				if e != nil || n > 65535 || n <= 0 {
					log.Fatal("PORT is not a valid number/not between 1-65535.")
				}
			}
			break
		case MasterKey, Endpoint, MongoConnectionURI, MongoDatabase, StandardKey:
			if os.Getenv(v) == "" {
				log.Fatal(fmt.Sprintf("%s must be set.", v))
			}
			break
		case MaxSizeBytes:
			if os.Getenv(v) == "" {
				fmt.Println("Setting " + v + " to 50 MiB")
				e := os.Setenv(v, "52428800")
				if e != nil {
					log.Fatal(fmt.Sprintf("Could not set %s with a default size", e))
				}
			} else {
				n, e := strconv.ParseInt(os.Getenv(v), 0, 64)
				if e != nil || n <= 0 {
					log.Fatal(fmt.Sprintf("%s is not greater than 0 B, or the syntax is invalid.", v))
				}
			}
			break
		}
	}
}