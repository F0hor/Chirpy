package main

import(
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"slices"
	"time"
	"fmt"

	"github.com/google/uuid"

	"github.com/F0hor/Chirpy/internal/database"
	"github.com/F0hor/Chirpy/internal/auth"
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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Missing validation token")
		return
	}

	tokenID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Validation error: %s", err)
		respondWithError(w, 401, "Broken or invalit validation token")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cenStr := censoreProfane(params.Body)
	if err != nil {
		log.Printf("Error while creating chirp: %s", err)
		respondWithError(w, 400, "Invalid user id")
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
		Pass string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	hash, err := auth.HashPassword(params.Pass)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.db.CreateUser(
		r.Context(), 
		database.CreateUserParams{
			Email: params.Email,
			HashedPassword: hash,
		},
	)
	if err != nil {
		log.Printf("Error creating user in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	respondWithJSON(w, 201, mapDbUser(user))
}

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass string `json:"password"`
		ExpiresIn int `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.db.GetUserByMail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error getting user in DB: %s", err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(params.Pass, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	ret := mapDbUser(user)
	expiresIn, err := time.ParseDuration("1h")
	if params.ExpiresIn > 0 && params.ExpiresIn < 3600 {
		expiresIn, err = time.ParseDuration(fmt.Sprintf("%vs", params.ExpiresIn))
	}

	token, err := auth.MakeJWT(ret.ID, cfg.secret, expiresIn)
	if err != nil {
		respondWithError(w, 500, "Failed to make validation token")
	}
	ret.Token = token

	respondWithJSON(w, 200, ret)
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

	respondWithJSON(w, 200, ret)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp id")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			respondWithError(w, 404, "No valid chirp")
			return
		}

		log.Printf("Error geting chirp in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	respondWithJSON(w, 200, mapDbChirp(chirp))
}

