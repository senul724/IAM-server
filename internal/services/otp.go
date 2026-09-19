package services

import (
	"IAM-server/internal/connections"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/redis/go-redis/v9"
)

type OTPReason string

const (
	OTPReasonLogin OTPReason = "login"
	OTPReasonReset OTPReason = "reset"
)

func otpKey(reason OTPReason, email string) string {
	return fmt.Sprintf("%s_%s", reason, email)
}

// GenerateOTP generates a 6-digit OTP and stores it in Redis with the given reason prefix
func GenerateOTP(ctx context.Context, email string, reason OTPReason) (string, error) {
	if connections.Redis == nil {
		return "", errors.New("redis connection is not initialized")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	otp := fmt.Sprintf("%06d", n.Int64()+100000)

	key := otpKey(reason, email)
	err = connections.Redis.Set(ctx, key, otp, OTP_TTL).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save OTP in redis: %w", err)
	}

	return otp, nil
}

// ValidateOTP validates the OTP crosschecking with Redis
func ValidateOTP(ctx context.Context, email string, otp string, reason OTPReason) error {
	if connections.Redis == nil {
		return errors.New("redis connection is not initialized")
	}

	key := otpKey(reason, email)
	storedOTP, err := connections.Redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("OTP expired or not found")
		}
		return fmt.Errorf("failed to get OTP from redis: %w", err)
	}

	if storedOTP != otp {
		return errors.New("invalid OTP")
	}

	_ = connections.Redis.Del(ctx, key).Err()

	return nil
}