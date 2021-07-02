package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func initRedisClient() (*redis.Client, func()) {
	client := redis.NewClient(&redis.Options{})
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		panic(err)
	}

	log.Println("redis connected")
	return client, func() {
		err := client.Close()
		if err != nil {
			log.Printf("close redis connection failed: %v\n", err)
		} else {
			log.Println("redis connection closed")
		}
	}
}
