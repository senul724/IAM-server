package tokens

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const ACCESS_TOKEN_TTL = 30 * time.Minute

type AccessTokenClaims struct {
	Email string `json:"email"`

	jwt.RegisteredClaims
}

func GenerateAccessToken(
	email string,
	userID string,
) (string, error) {

	claims := AccessTokenClaims{
		Email: email,

		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ACCESS_TOKEN_TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func VerifyAccessToken(tokenString string) (*AccessTokenClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessTokenClaims{},
		func(token *jwt.Token) (any, error) {

			// Ensure the expected signing method was used
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("ACCESS_SECRET")), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
