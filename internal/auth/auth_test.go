package auth

import (
  "testing"
	"time"
	"fmt"

	"github.com/google/uuid"
)

func TestHelloName(t *testing.T) {
	userId := uuid.New()
	secret := "aaaa"

	expiresIn, err := time.ParseDuration("1h")
	if err != nil {
		fmt.Println(err)
	}

	ss, err := MakeJWT(userId, secret, expiresIn)
	if err != nil {
		fmt.Println("make - ", err)
	}

	newId, err := ValidateJWT(ss, secret)
	if err != nil {
		fmt.Println("val - ", err)
	}

	if userId != newId {
		t.Errorf("New ID != user Id")
	}
}
