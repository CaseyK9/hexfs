package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func SendJSONResponse(w *http.ResponseWriter, i interface{}) {
	(*w).Header().Set("Content-Type", "application/json")
	sendErr := json.NewEncoder(*w).Encode(i)
	if sendErr != nil {
		fmt.Println("Error sending JSON to client: "  + sendErr.Error())
	}
	return
}
