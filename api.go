package main

import(
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"slices"
)

var profaneWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	type returnVals struct {
		Body string `json:"cleaned_body"`
	}
	ret := returnVals{
		Body: censoreProfane(params.Body),
	}

	respondWithJSON(w, 200, ret)
}

func censoreProfane(txt string) string {
	words := strings.Split(txt, " ")

	for i, w := range words {
		if slices.Contains(profaneWords, strings.ToLower(w)) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

