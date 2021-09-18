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

func SaveToCache(key string, value string, timeout int) {
	conn := RedisClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := conn.Set(ctx, key, value, time.Duration(timeout)*time.Minute).Err()
	if err != nil {
		log.Println("Unable to save to cache")
	} else {
		log.Println("Saved to cache!")
	}
}

func IsPresentInCache(key string) (value string, e error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client := RedisClient()
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Key %s was not present in cache\n", key)
		return "", err
	} else if err != nil {
		log.Printf("Unable to interact with cache for key %s\n", key)
		return "", err
	} else {
		return val, err
	}
}
