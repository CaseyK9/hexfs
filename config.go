package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

const (
	Port = "PSE_PORT"
	UploadKey = "PSE_UPLOAD_KEY"
	MinSizeBytes = "PSE_MIN_SIZE_BYTES"
	MaxSizeBytes = "PSE_MAX_SIZE_BYTES"
	DiscordWebhookURL = "PSE_DISCORD_WEBHOOK"
	UploadDirMaxSize = "PSE_UPLOAD_DIR_MAX_SIZE"
	UploadDirPath = "PSE_UPLOAD_DIR_PATH"
	Endpoint = "PSE_ENDPOINT"
	Frontend = "PSE_FRONTEND"
)

func ValidateEnv() {
	for _, v := range []string{
		Port,
		UploadKey,
		MinSizeBytes,
		MaxSizeBytes,
		UploadDirMaxSize,
		UploadDirPath,
	} {
		switch v {
		case UploadDirPath:
			if os.Getenv(v) == "" {
				p, pathErr := os.Getwd()
				if pathErr != nil {
					panic("Cannot determine current working directory.")
				}
				fmt.Println("Setting uploads folder to " + path.Join(p, "uploads"))
				e := os.Setenv(v, path.Join(p, "uploads"))
				if e != nil {
					panic("Could not set default upload directory path.")
				}
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
		case UploadKey:
			if os.Getenv(v) == "" {
				panic(fmt.Sprintf("%s must be set.", v))
			}
			break
		case MinSizeBytes, MaxSizeBytes, UploadDirMaxSize:
			if os.Getenv(v) == "" {
				switch v {
				case MinSizeBytes:
					fmt.Println("Setting " + v + " to 512 B")
					e := os.Setenv(v, "512")
					if e != nil {
						panic(fmt.Sprintf("Could not set %s with a default size", e))
					}
					break
				case MaxSizeBytes:
					fmt.Println("Setting " + v + " to 50 MiB")
					e := os.Setenv(v, "52428800")
					if e != nil {
						panic(fmt.Sprintf("Could not set %s with a default size", e))
					}
					break
				case UploadDirMaxSize:
					fmt.Println("Setting " + v + " to 10 GiB")
					e := os.Setenv(UploadDirMaxSize, "10737418240")
					if e != nil {
						panic("Could not set " + UploadDirMaxSize)
					}
					break
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