package main

import(
	"net/http"

	"github.com/google/uuid"

	"github.com/F0hor/Chirpy/internal/jsonhand"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type data struct {
		UserID string `json:"user_id"`
	}
	type parameters struct {
		Event string `json:"event"`
		Data data `json:"data"`
	}

	params := parameters{}
	err := jsonhand.Decode(w, r, &params)
	if err != nil {
		return
	}

	switch params.Event {
    case "user.upgraded":
      cfg.PolkaUpgradeUser(w, r, params.Data.UserID)
			return
  }

	w.WriteHeader(204)
}

func (cfg *apiConfig) PolkaUpgradeUser(w http.ResponseWriter, r *http.Request, userIDstr string) {
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		jsonhand.RespondWithError(w, 404, "Failed to parse user ID")
		return
	}

	_, err = cfg.db.UpgradeUser(r.Context(), userID)
	if err != nil {
		jsonhand.RespondWithError(w, 404, "User not found")
		return
	}

	w.WriteHeader(204)
}

