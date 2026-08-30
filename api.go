package main

import(
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"slices"

	"github.com/google/uuid"

	"github.com/F0hor/Chirpy/internal/database"
)

var profaneWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		UserID string `json:"user_id"`
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

	cenStr := censoreProfane(params.Body)
	uid, err := uuid.Parse(params.UserID)
	if err != nil {
		log.Printf("Error while creating chirp: %s", err)
		respondWithError(w, 400, "Invalid user id")
		return
	}

	chirp, err := cfg.db.CreateChirp(
		r.Context(),
		database.CreateChirpParams{
			Body: cenStr,
			UserID: uid,
		},
	)
	if err != nil {
		log.Printf("Error creating chirp in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	respondWithJSON(w, 201, mapDbChirp(chirp))
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

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating user in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	respondWithJSON(w, 201, mapDbUser(user))
}

