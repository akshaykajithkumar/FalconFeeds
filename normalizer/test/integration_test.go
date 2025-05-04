package normalizer

import (
	"context"
	"testing"
	"time"

	"normalizer/internal/normalizer"
	"normalizer/internal/shared"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

func TestProcessorEndToEndFlow(t *testing.T) {
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("MONGO_HOST", "localhost")
	t.Setenv("MONGO_PORT", "27017")

	redisClient := shared.NewRedisClient()
	mongoClient, mongoColl := shared.NewMongoCollection()
	defer mongoClient.Disconnect(context.Background())

	// Clear collection with timeout
	ctxDelete, cancelDelete := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDelete()
	_, err := mongoColl.DeleteMany(ctxDelete, bson.M{})
	if err != nil {
		t.Fatalf("Could not clear database: %v", err)
	}

	processor := normalizer.NewProcessor(redisClient, mongoColl)

	// Add test payload
	ctxAdd, cancelAdd := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelAdd()
	_, err = redisClient.XAdd(ctxAdd, &redis.XAddArgs{
		Stream: "raw-feeds",
		Values: map[string]interface{}{"payload": "Suspicious activity from 1.2.3.4"},
	}).Result()
	if err != nil {
		t.Fatalf("Failed to add Redis message: %v", err)
	}

	// Run processor with timeout
	ctxRun, cancelRun := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelRun()
	go processor.Start(ctxRun)

	// Poll MongoDB instead of fixed sleep
	var found bool
	for start := time.Now(); time.Since(start) < 10*time.Second; {
		ctxFind, cancelFind := context.WithTimeout(context.Background(), 2*time.Second)
		count, _ := mongoColl.CountDocuments(ctxFind, bson.M{})
		cancelFind()

		if count > 0 {
			found = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !found {
		t.Error("No documents found in MongoDB after processing")
	}
}
