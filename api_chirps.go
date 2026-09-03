package main

import(
	"net/http"
	"log"
	"strings"
	"slices"

	"github.com/google/uuid"

	"github.com/F0hor/Chirpy/internal/database"
	"github.com/F0hor/Chirpy/internal/auth"
	"github.com/F0hor/Chirpy/internal/jsonhand"
)

var profaneWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}
	err := jsonhand.Decode(w, r, params)
	if err != nil {
		return
	}

	tokenID, err := auth.GetUserIDFromHeader(w, r.Header, cfg.secret)
	if err != nil {
		return
	}

	if len(params.Body) > 140 {
		jsonhand.RespondWithError(w, 400, "Chirp is too long")
		return
	}

	cenStr := censoreProfane(params.Body)
	if err != nil {
		log.Printf("Error while creating chirp: %s", err)
		jsonhand.RespondWithError(w, 400, "Invalid user id")
		return
	}

	chirp, err := cfg.db.CreateChirp(
		r.Context(),
		database.CreateChirpParams{
			Body: cenStr,
			UserID: tokenID,
		},
	)
	if err != nil {
		log.Printf("Error creating chirp in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	jsonhand.RespondWithJSON(w, 201, mapDbChirp(chirp))
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

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error geting chirps in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	ret := []Chirp{}
	for _, c := range chirps {
		ret = append(ret, mapDbChirp(c))
	}

	jsonhand.RespondWithJSON(w, 200, ret)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		jsonhand.RespondWithError(w, 400, "Invalid chirp id")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			jsonhand.RespondWithError(w, 404, "No valid chirp")
			return
		}

		log.Printf("Error geting chirp in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	jsonhand.RespondWithJSON(w, 200, mapDbChirp(chirp))
}

