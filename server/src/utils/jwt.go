package utils

import (
	"IAM-server/src/types"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const key = "yqY9OPPUy4RouGWbelqwUwlxqyu9NwzFMZNrZJlcfLV"

func CreateRefreshToken(userId string, userdata *types.UserData) (string, error) {
	jsonData, err := json.Marshal(userdata)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":  uuid.New(),
		"sub":  userId,
		"user": jsonData,
		"exp":  time.Now().Add(time.Minute).Unix(),
	})

	return token.SignedString([]byte(key))
}
