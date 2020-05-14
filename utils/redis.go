package utils

import (
	"github.com/go-redis/redis/v7"
	"os"
)

var client *redis.Client

func init() {
	//Initializing redis
	redisAddr := os.Getenv("REDIS_ADDRESS")
	redisPass := os.Getenv("REDIS_PASSWORD")

	if len(redisAddr) == 0 {
		redisAddr = "localhost:6379"
	}

	if len(redisPass) == 0 {
		redisPass = ""
	}

	client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
}
