package tokens

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SessionTokenClaims struct {
	Name  string `json:"name"`
	Email string `json:"email"`

	jwt.RegisteredClaims
}

const SESSION_TOKEN_TTL = 15 * 24 * time.Hour

func GenerateSessionToken(
	name string,
	email string,
	userID string,
) (string, error) {

	claims := SessionTokenClaims{
		Name:  name,
		Email: email,

		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(SESSION_TOKEN_TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString([]byte(os.Getenv("SESSION_SECRET")))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func VerifySessionToken(tokenString string) (*SessionTokenClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&SessionTokenClaims{},
		func(token *jwt.Token) (any, error) {

			// Ensure the expected signing method was used
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("SESSION_SECRET")), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SessionTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
