package utils

import (
	"IAM-server/src/types"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	Key        = "yqY9OPPUy4RouGWbelqwUwlxqyu9NwzFMZNrZJlcfLV"
	refreshExp = time.Hour * 24 * 14 // 14 days
	accessExp  = time.Minute * 5     // 14 days
)

func CreateRefreshToken(userId string, userdata types.UserData) (string, error) {
	return createJwt(userId, userdata, refreshExp, Key)
}

func CreateAccessToken(userId string, userdata types.UserData) (string, error) {
	return createJwt(userId, userdata, accessExp, Key)
}

func createJwt(userId string, userdata types.UserData, exp time.Duration, key string) (string, error) {
	claims := types.CustomClaims{
		User: userdata,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://localhost:8000",
			Subject:   userId,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

func VerifyRefreshToken(r *http.Request) error {
	cookie, cookieErr := r.Cookie("iam-refresh")
	if cookieErr != nil {
		return cookieErr
	}
	_, validationError := validateToken(cookie.Value, Key)
	return validationError
}

func VerifyAccessToken(r *http.Request) (*types.CustomClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("Authorization header not present")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return nil, fmt.Errorf("Invalid Authorization Header")
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	return validateToken(tokenString, Key)
}

func validateToken(tokenString string, secretKey string) (*types.CustomClaims, error) {
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
