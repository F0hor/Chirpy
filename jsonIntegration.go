package main

import(
	"net/http"
	"encoding/json"
	"log"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnError struct {
		ErrorMsg string `json:"error"`
	}

	respBody := returnError{
		ErrorMsg: msg,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}

	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}

	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  w.Write(dat)
}

