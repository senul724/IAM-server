package tokens

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const REFRESH_TOKEN_TTL = 15 * 24 * time.Hour

func GenerateRefreshToken(
	userID string,
) (string, error) {

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(REFRESH_TOKEN_TTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString([]byte(os.Getenv("REFRESH_SECRET")))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func VerifyRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {

			// Ensure the expected signing method was used
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("REFRESH_SECRET")), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
