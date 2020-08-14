package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"path"
)

func ServeDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authErr := IsAuthorized(r)
	if authErr != nil {
		SendJSONResponse(&w, authErr)
		return
	}

	fPath := path.Join(os.Getenv(UploadDirPath), ps.ByName("name"))
	_, statErr := os.Stat(fPath)
	if statErr != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "File ID does not exist.",
		})
		return
	}
	sizeOfDir, _ := DirSize(fPath)
	err := os.RemoveAll(fPath)
	if err != nil {
		SendJSONResponse(&w, ResponseError{
			Status:  1,
			Message: "Failed to delete file. " + err.Error(),
		})
		return
	}
	if os.Getenv(DiscordWebhookURL) != "" {
		webhookErr := SendToWebhook(fmt.Sprintf("File ID %s deleted. Freed **%s** of space.", ps.ByName("name"), ByteCountSI(uint64(sizeOfDir))))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}
	SizeOfUploadDir -= sizeOfDir
	SendJSONResponse(&w, EmptyResponse{
		Status:  0,
	})
	return
}
