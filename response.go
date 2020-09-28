package main

import (
	json2 "encoding/json"
	"fmt"
	"log"
	"net/http"
)

// SendTextResponse sends a plaintext response to the client along with an HTTP status code.
func SendTextResponse(w *http.ResponseWriter, msg string, code int) {
	if code == http.StatusInternalServerError {
		log.Printf(fmt.Sprintf("Unhandled error!, %s", msg))
	}
	(*w).WriteHeader(code)
	_, _ = fmt.Fprintln(*w, msg)
	return
}

// SendJSONResponse sends a JSON encoded response to the client along with an HTTP status code of 200 OK.
func SendJSONResponse(w *http.ResponseWriter, json interface{}) {
	(*w).Header().Set("Content-Type", "application/json")
	_ = json2.NewEncoder(*w).Encode(json)
}

// SendNothing sends 204 No Content.
func SendNothing(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusNoContent)
	return
}