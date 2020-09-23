package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"google.golang.org/api/option"
	"os"
)

func CreateGCSClient() (*storage.Client, error) {
	c, err := storage.NewClient(context.Background(), option.WithCredentialsFile(os.Getenv(GoogleApplicationCredentials)))
	if err != nil {
		return nil, err
	}
	return c, nil
}


func GetKey() ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(os.Getenv(GCSSecretKey))
	if err != nil {
		return nil, err
	}
	return key, nil
}