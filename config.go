package main

import (
	"github.com/vysiondev/hexfs/hlog"
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
	GCSBucketName = "GCS_BUCKET_NAME"
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	GCSSecretKey = "GCS_SECRET_KEY"
	DisableFileBlacklist = "HFS_DISABLE_FILE_BLACKLIST"
	MongoConnectionURI = "HFS_MONGO_CONNECTION_URI"
	MongoDatabase = "HFS_MONGO_DATABASE"
	ContainerNickname = "HFS_CONTAINER_NICKNAME"
	FaviconLocation = "HFS_FAVICON_LOCATION"
	RedisConnectionURI = "HFS_REDIS_CONNECTION_URI"
	RedisPassword = "HFS_REDIS_PASSWORD"
	RedisDbInt = "HFS_REDIS_DB"
	MaxCapacity = "HFS_MAX_CAPACITY"
	Port = "HFS_PORT"
)

// ValidateEnv validates the environment variables and throws log.Fatal if a variable is not correctly set or not set at all.
func ValidateEnv() {
	for _, v := range []string{
		MasterKey,
		StandardKey,
		MaxSizeBytes,
		GCSBucketName,
		GoogleApplicationCredentials,
		GCSSecretKey,
		MongoConnectionURI,
		MongoDatabase,
		ContainerNickname,
		RedisConnectionURI,
		RedisPassword,
		MaxCapacity,
		Port,
	} {
		if len(os.Getenv(v)) == 0 {
			switch v {
			case MasterKey:
			case StandardKey:
			case GCSBucketName:
			case GCSSecretKey:
			case GoogleApplicationCredentials:
			case MongoConnectionURI:
			case MongoDatabase:
			case RedisConnectionURI:
			case MaxCapacity:
				missing(v)
				break
			case MaxSizeBytes:
				hlog.Log("config", hlog.LevelInfo, "Setting max size of files to 50 MiB because it was not set")
				e := os.Setenv(MaxSizeBytes, "52428800")
				if e != nil { cannotSet(v) }
				break
			case ContainerNickname:
				n := "hexfs_" + strconv.FormatInt(time.Now().Unix(), 10)
				hlog.Log("config", hlog.LevelInfo, "Setting default container nickname to " + n + " because it was not set")
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