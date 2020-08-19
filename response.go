package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func SendJSONResponse(w *http.ResponseWriter, i interface{}, code int) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(code)
	sendErr := json.NewEncoder(*w).Encode(i)
	if sendErr != nil {
		fmt.Println("There was a problem sending JSON to client: "  + sendErr.Error())
		(*w).WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(*w, "{\"status\": 1, \"message\": \"Failed to encode JSON response. %s\"}", sendErr.Error())
	}
	return
}

func SendNothing(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusNoContent)
	return
}
