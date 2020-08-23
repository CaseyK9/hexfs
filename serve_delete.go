package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"path"
)

func ServeDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authSuccess := IsAuthorized(w, r)
	if !authSuccess {
		return
	}

	fPath := path.Join(os.Getenv(UploadDirPath), ps.ByName("name"))
	_, statErr := os.Stat(fPath)
	if statErr != nil {
		SendTextResponse(&w, "File ID does not exist.", http.StatusNotFound)
		return
	}
	sizeOfDir, _ := DirSize(fPath)
	err := os.RemoveAll(fPath)
	if err != nil {
		SendTextResponse(&w, "Failed to delete file. " + err.Error(), http.StatusInternalServerError)
		return
	}
	if os.Getenv(DiscordWebhookURL) != "" {
		webhookErr := SendToWebhook(fmt.Sprintf("File ID %s deleted. Freed **%s** of space.", ps.ByName("name"), ByteCountSI(uint64(sizeOfDir))))
		if webhookErr != nil {
			fmt.Println("Webhook failed to send: " + webhookErr.Error())
		}
	}
	SendNothing(&w)
	return
}
