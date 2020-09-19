package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	Port = "HFS_PORT"
	UploadKey = "HFS_UPLOAD_KEY"
	MaxSizeBytes = "HFS_MAX_SIZE_BYTES"
	Endpoint = "HFS_ENDPOINT"
	Frontend = "HFS_FRONTEND"
	GCSBucketName = "GCS_BUCKET_NAME"
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	GCSSecretKey = "GCS_SECRET_KEY"
)

func ValidateEnv() {
	for _, v := range []string{
		Port,
		UploadKey,
		MaxSizeBytes,
		Endpoint,
		GCSBucketName,
		GoogleApplicationCredentials,
	} {
		switch v {
		case GCSBucketName:
		case GoogleApplicationCredentials:
		case GCSSecretKey:
			if os.Getenv(v) == "" {
				panic(fmt.Sprintf("You must set the proper Google Cloud Storage variables."))
			}
		case Port:
			if os.Getenv(v) == "" {
				e := os.Setenv(v, "7250")
				if e != nil {
					panic("Could not set default port to 7250")
				}
			} else {
				n, e := strconv.ParseInt(os.Getenv(v), 0, 64)
				if e != nil || n > 65535 || n <= 0 {
					panic("PORT is not a valid number/not between 1-65535.")
				}
			}
			break
		case UploadKey, Endpoint:
			if os.Getenv(v) == "" {
				panic(fmt.Sprintf("%s must be set.", v))
			}
			break
		case MaxSizeBytes:
			if os.Getenv(v) == "" {
				fmt.Println("Setting " + v + " to 50 MiB")
				e := os.Setenv(v, "52428800")
				if e != nil {
					panic(fmt.Sprintf("Could not set %s with a default size", e))
				}
			} else {
				n, e := strconv.ParseInt(os.Getenv(v), 0, 64)
				if e != nil || n <= 0 {
					panic(fmt.Sprintf("%s is not greater than 0 B, or the syntax is invalid.", v))
				}
			}
			break
		}
	}
}