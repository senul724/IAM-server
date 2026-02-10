package utils

import (
	"IAM-server/src/types"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	Key            = "yqY9OPPUy4RouGWbelqwUwlxqyu9NwzFMZNrZJlcfLV"
	expirationTime = time.Hour * 24 * 14 // 14 days
)

func CreateRefreshToken(userId string, userdata types.UserData) (string, error) {
	claims := types.CustomClaims{
		User: userdata,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://localhost:8000",
			Subject:   userId,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(Key))
}

func CheckRefreshToken(r *http.Request) error {
	cookie, cookieErr := r.Cookie("iam-refresh")
	if cookieErr != nil {
		return cookieErr
	}
	_, validationError := validateTokenWithClaims(cookie.Value, Key)
	return validationError
}

func validateToken(tokenString string, secretKey string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func validateTokenWithClaims(tokenString string, secretKey string) (*types.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*types.CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
