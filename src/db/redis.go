package db

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func RedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}

func SaveUserResponse(key string, value string) {
	conn := RedisClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := conn.Set(ctx, key, value, 30*time.Minute).Err()
	if err != nil {
		log.Println("Unable to save to cache")
	} else {
		log.Println("Saved to cache!")
	}
}
