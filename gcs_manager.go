package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"google.golang.org/api/option"
	"os"
)

func CreateGCSClient() (*storage.Client, context.Context, error) {
	ctx := context.Background()
	c, err := storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv(GoogleApplicationCredentials)))
	if err != nil {
		return nil, nil, err
	}
	return c, ctx, nil
}


func GetKey() ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(os.Getenv(GCSSecretKey))
	if err != nil {
		return nil, err
	}
	return key, nil
}