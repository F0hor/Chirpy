package main

import(
	"time"
	
	"github.com/google/uuid"

	"github.com/F0hor/Chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token			string		`json:"token"`
}

func mapDbUser(u database.User) User {
	return User{
		ID: u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email: u.Email,
	}
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body     	string    `json:"body"`
	UserID		uuid.UUID	`json:"user_id"`
}

func mapDbChirp(c database.Chirp) Chirp {
	return Chirp{
		ID: c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body: c.Body,
		UserID: c.UserID,
	}
}

