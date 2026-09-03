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
	"github.com/F0hor/Chirpy/internal/jsonHand"
)

var profaneWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass string `json:"password"`
	}

	params := parameters{}
	err := decode(w, r, &params)
	if err != nil {
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

	jsonHand.respondWithJSON(w, 201, mapDbUser(user))
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass string `json:"password"`
	}

	params := parameters{}
	err := decode(w, r, &params)
	if err != nil {
		return
	}

	tokenId, err := GetUserIDFromHeader(w, r.Header, cfg.secret)
	if err != nil {
		return
	}

	hash, err := auth.HashPassword(params.Pass)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.db.UpdateUser(
		r.Context(),
		database.UpdateUserParams{
			ID: tokenId,
			Email: params.Email,
			HashedPassword: hash,
		},
	)
	if err != nil {
		log.Printf("Failed to update user in DB: %s", err)
		w.WriteHeader(500)
		return
	}

	jsonHand.respondWithJSON(w, 200, mapDbUser(user))	
}

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass string `json:"password"`
		ExpiresIn int `json:"expires_in_seconds"`
	}

	params := parameters{}
	err := decode(w, r, &params)
	if err != nil {
		return
	}

	user, err := cfg.db.GetUserByMail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error getting user in DB: %s", err)
		jsonHand.respondWithError(w, 401, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(params.Pass, user.HashedPassword)
	if err != nil || !match {
		jsonHand.respondWithError(w, 401, "Incorrect email or password")
		return
	}

	ret := mapDbUser(user)
	expiresIn, err := time.ParseDuration("1h")
	if params.ExpiresIn > 0 && params.ExpiresIn < 3600 {
		expiresIn, err = time.ParseDuration(fmt.Sprintf("%vs", params.ExpiresIn))
	}

	token, err := auth.MakeJWT(ret.ID, cfg.secret, expiresIn)
	if err != nil {
		jsonHand.respondWithError(w, 500, "Failed to make validation token")
	}
	ret.Token = token

	refresh, err := cfg.db.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			Token: auth.MakeRefreshToken(),
			UserID: ret.ID,
		},
	)
	if err != nil {
		jsonHand.respondWithError(w, 500, "Failed to make validation token")
	}
	ret.Refresh = refresh.Token

	jsonHand.respondWithJSON(w, 200, ret)
}

func (cfg *apiConfig) handlerRefreshUser(w http.ResponseWriter, r *http.Request) {
	bToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonHand.respondWithError(w, 401, "Missing or invalid refresh token")
		return
	}

	refresh, err := cfg.db.GetRefreshToken(r.Context(), bToken)
	if err != nil {
		jsonHand.respondWithError(w, 401, "Missing or invalid refresh token")
		return
	}

	if refresh.RevokedAt.Valid || time.Now().After(refresh.ExpiresAt) {
		jsonHand.respondWithError(w, 401, "Missing or invalid refresh token")
		return
	}

	expiresIn, err := time.ParseDuration("1h")
	token, err := auth.MakeJWT(refresh.UserID, cfg.secret, expiresIn)
	if err != nil {
		jsonHand.respondWithError(w, 500, "Failed to make validation token")
		return
	}

	type ref struct {
		Token string `json:"token"`
	}
	jsonHand.respondWithJSON(w, 200, ref{Token: token})
}

func (cfg *apiConfig) handlerRevokeUser(w http.ResponseWriter, r *http.Request) {
	bToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonHand.respondWithError(w, 401, "Missing or invalid refresh token")
		return
	}

	err = cfg.db.RevokeRefresh(r.Context(), bToken)
	if err != nil {
		jsonHand.respondWithError(w, 401, "Missing or invalid refresh token")
		return
	}

	w.WriteHeader(204)
}

