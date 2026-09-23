package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/utils/tokens"
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func RefreshJTIKey(jti string) string {
	return "auth:refresh:" + jti
}

func AccessJTIKey(jti string) string {
	return "auth:access:" + jti
}

// StoreTokenPair stores the refresh_jti -> access_jti and access_jti -> userID mappings in Redis.
func StoreTokenPair(ctx context.Context, refreshJTI, accessJTI, userID string) error {
	if connections.Redis == nil {
		return errors.New("redis connection is not initialized")
	}

	pipe := connections.Redis.Pipeline()
	pipe.Set(ctx, RefreshJTIKey(refreshJTI), accessJTI, tokens.REFRESH_TOKEN_TTL)
	pipe.Set(ctx, AccessJTIKey(accessJTI), userID, tokens.ACCESS_TOKEN_TTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store token pair in redis: %w", err)
	}

	return nil
}

// ValidateAccessTokenJTI checks if the access JTI exists in Redis and returns the associated userID.
func ValidateAccessTokenJTI(ctx context.Context, accessJTI string) (string, error) {
	if connections.Redis == nil {
		return "", errors.New("redis connection is not initialized")
	}

	key := AccessJTIKey(accessJTI)
	userID, err := connections.Redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			rawVal, rawErr := connections.Redis.Get(ctx, accessJTI).Result()
			if rawErr == nil {
				return rawVal, nil
			}
			return "", errors.New("access token revoked or expired")
		}
		return "", fmt.Errorf("failed to validate access token in redis: %w", err)
	}

	return userID, nil
}

// ValidateRefreshTokenJTI checks if the refresh JTI exists in Redis and returns the associated accessJTI.
func ValidateRefreshTokenJTI(ctx context.Context, refreshJTI string) (string, error) {
	if connections.Redis == nil {
		return "", errors.New("redis connection is not initialized")
	}

	key := RefreshJTIKey(refreshJTI)
	accessJTI, err := connections.Redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			rawVal, rawErr := connections.Redis.Get(ctx, refreshJTI).Result()
			if rawErr == nil {
				return rawVal, nil
			}
			return "", errors.New("refresh token revoked or expired")
		}
		return "", fmt.Errorf("failed to validate refresh token in redis: %w", err)
	}

	return accessJTI, nil
}

// RevokeTokenPairByRefreshJTI finds the associated accessJTI from refreshJTI and deletes both records from Redis.
func RevokeTokenPairByRefreshJTI(ctx context.Context, refreshJTI string) error {
	if connections.Redis == nil {
		return errors.New("redis connection is not initialized")
	}

	accessJTI, _ := ValidateRefreshTokenJTI(ctx, refreshJTI)

	pipe := connections.Redis.Pipeline()
	pipe.Del(ctx, RefreshJTIKey(refreshJTI))
	pipe.Del(ctx, refreshJTI)
	if accessJTI != "" {
		pipe.Del(ctx, AccessJTIKey(accessJTI))
		pipe.Del(ctx, accessJTI)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke token pair from redis: %w", err)
	}

	return nil
}

// RevokeAccessTokenJTI deletes the access JTI from Redis.
func RevokeAccessTokenJTI(ctx context.Context, accessJTI string) error {
	if connections.Redis == nil {
		return errors.New("redis connection is not initialized")
	}

	pipe := connections.Redis.Pipeline()
	pipe.Del(ctx, AccessJTIKey(accessJTI))
	pipe.Del(ctx, accessJTI)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke access token from redis: %w", err)
	}

	return nil
}
