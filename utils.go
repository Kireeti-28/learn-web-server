package main

import (
	"encoding/json"
	"net/http"
)

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(msg))
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data any) {
	dat, _ := json.Marshal(data)
	w.WriteHeader(statusCode)
	w.Write(dat)
}
