package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	Port = "HFS_PORT"
	UploadKey = "HFS_UPLOAD_KEY"
	DeletionKey = "HFS_DELETION_KEY"
	MaxSizeBytes = "HFS_MAX_SIZE_BYTES"
	Endpoint = "HFS_ENDPOINT"
	Frontend = "HFS_FRONTEND"
	GCSBucketName = "GCS_BUCKET_NAME"
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	GCSSecretKey = "GCS_SECRET_KEY"
	PublicMode = "HFS_PUBLIC_MODE"
	DisableFileBlacklist = "HFS_DISABLE_FILE_BLACKLIST"
)

func ValidateEnv() {
	for _, v := range []string{
		Port,
		UploadKey,
		DeletionKey,
		PublicMode,
		MaxSizeBytes,
		Endpoint,
		GCSBucketName,
		GoogleApplicationCredentials,
		DisableFileBlacklist,
	} {
		switch v {
		case PublicMode, DisableFileBlacklist:
			if os.Getenv(v) == "" || os.Getenv(v) != "1" {
				e := os.Setenv(v, "0")
				if e != nil {
					panic("Default value of " + v + " could not be set to 0.")
				}
			} else if os.Getenv(v) == "1" {
				if v == PublicMode {
					fmt.Println("!!!! WARNING! Public mode ENABLED. Anonymous uploading is allowed! !!!!")
				} else {
					fmt.Println("!!!! WARNING! File blacklist is DISABLED. Malicious files can be uploaded! !!!!")
				}
			}
			break
		case GCSBucketName, GoogleApplicationCredentials, GCSSecretKey:
			if os.Getenv(v) == "" {
				panic(fmt.Sprintf("You must set the proper Google Cloud Storage variables."))
			}
			break
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