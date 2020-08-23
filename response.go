package main

import (
	"fmt"
	"net/http"
)

func SendTextResponse(w *http.ResponseWriter, msg string, code int) {
	(*w).WriteHeader(code)
	_, _ = fmt.Fprintln(*w, msg)
	return
}

func SendNothing(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusNoContent)
	return
}
