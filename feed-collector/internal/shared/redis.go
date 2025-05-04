package shared

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// Retry logic
	for i := 0; i < 10; i++ {
		err := client.Ping(ctx).Err()
		if err == nil {
			log.Println("Connected to Redis")
			return client
		}

		log.Printf("Waiting for Redis... (%v)", err)
		time.Sleep(2 * time.Second)
	}

	log.Fatal("Failed to connect to Redis after retries")
	return nil
}
