package auth

import(
	"log"
	"time"
	"errors"
	"strings"
	"net/http"
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"

	"github.com/F0hor/Chirpy/internal/jsonhand"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer: "chirpy-access",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject: userID.String(),
		},
	)
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString, 
		&jwt.RegisteredClaims{}, 
		func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},	
	)
	if err != nil {
		return uuid.New(), err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.New(), errors.New("Failed to retrieve claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.New(), err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	return GetAuthHeader(headers, 1)
}

func GetAPIKey(headers http.Header) (string, error) {
	return GetAuthHeader(headers, 1)
}

func GetAuthHeader(headers http.Header, elem int) (string, error) {
	bearer := headers.Get("Authorization")
	if bearer == "" {
		return "", errors.New("Missing Authorization")
	}

	return strings.Split(bearer, " ")[elem], nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func GetUserIDFromHeader(w http.ResponseWriter, headers http.Header, tokenSecret string) (uuid.UUID, error) {
	token, err := GetBearerToken(headers)
	if err != nil {
		jsonhand.RespondWithError(w, 401, "Missing validation token")
		return uuid.New(), err
	}

	tokenID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		log.Printf("Validation error: %s", err)
		jsonhand.RespondWithError(w, 401, "Broken or invalit validation token")
		return uuid.New(), err
	}

	return tokenID, err
}

