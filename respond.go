package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func SendJSONResponse(w *http.ResponseWriter, i interface{}) {
	sendErr := json.NewEncoder(*w).Encode(i)
	if sendErr != nil {
		fmt.Println("Error sending JSON to client: "  + sendErr.Error())
	}
	return
}
