package jsonhand

import(
	"net/http"
	"encoding/json"
	"log"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
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

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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

func Decode(w http.ResponseWriter, r *http.Request, params any) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		RespondWithError(w, 500, "Failed to decode request body")
		return err
	}
	return nil
}
