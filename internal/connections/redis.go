package connections

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis() {
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		log.Println("Warning: REDIS_URL is not set in environment")
		return
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: failed to connect to Redis: %v", err)
	} else {
		log.Println("Redis connected successfully")
	}

	Redis = client
}
