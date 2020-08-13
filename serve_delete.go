package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"path"
)

func ServeDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ApplyCORSHeaders(&w)
	auth := IsAuthorized(r)
	if !auth {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Not authorized to delete.",
		})
		return
	}
	if ps.ByName("name") == "404.png" {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "This file is reserved and cannot be deleted.",
		})
		return
	}
	fi, statErr := os.Stat(path.Join(os.Getenv(UploadDirPath), ps.ByName("name")))
	if statErr != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Failed to get information about file. " + statErr.Error(),
		})
		return
	}
	err := os.Remove(path.Join(os.Getenv(UploadDirPath), ps.ByName("name")))
	if err != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Failed to delete file. " + err.Error(),
		})
		return
	}
	if os.Getenv(DiscordWebhookURL) != "" {
		webhookErr := SendToWebhook(fmt.Sprintf("%s deleted. Freed **%s** of space.", ps.ByName("name"), ByteCountSI(uint64(fi.Size()))))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}
	SizeOfUploadDir -= fi.Size()
	SendJSONResponse(&w, EmptyResponse{
		Status:  0,
	})
	return
}
