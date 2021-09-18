package db

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const LongTTL = 10000000

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
	ctx := GetCTX()
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

//First, ensure keys are saved with absurdly long TTL
func IsPresentInCacheWithTime(key string) (value string, e error) {
	client := RedisClient()
	ttl, _ := client.TTL(GetCTX(), key).Result()
	if ttl < 0 { // Key must not exist/some other error
		return "", errors.New("key not in redis")
	} else {
		if (LongTTL - ttl.Minutes()) >= 0 {
			val, _ := IsPresentInCache(key)
			if val != "" {
				return val, errors.New("key past expiry - Mongo Query Needed")
			} else {
				return "", errors.New("key not in redis") // Might as well treat it as not present since fetch was unsuccessful
			}
		} else { //All OK
			return IsPresentInCache(key)
		}
	}
}
