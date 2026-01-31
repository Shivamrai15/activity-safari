package utils

import (
	"context"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

func InitRedisClient() {
	redisOnce.Do(func() {
		REDIS_URL := GetEnvValue("REDIS_URL", "", true)
		REDIS_PASSWORD := GetEnvValue("REDIS_PASSWORD", "", true)

		redisClient = redis.NewClient(&redis.Options{
			Addr:     REDIS_URL,
			Username: "default",
			Password: REDIS_PASSWORD,
			DB:       0,
		})

		ctx := context.Background()
		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			log.Fatal("Failed to connect to Redis:", err)
		}

		log.Println("Redis connection established successfully")
	})
}

func GetRedisClient() *redis.Client {
	if redisClient == nil {
		InitRedisClient()
	}
	return redisClient
}
