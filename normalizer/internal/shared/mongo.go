package shared

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// func NewMongoCollection() (*mongo.Client, *mongo.Collection) {
// 	ctx := context.Background()
// 	clientOpts := options.Client().ApplyURI("mongodb://mongo:27017")
// 	client, err := mongo.Connect(ctx, clientOpts)
// 	if err != nil {
// 		log.Fatalf("MongoDB connect error: %v", err)
// 	}

//		return client, client.Database("falconfeeds").Collection("normalized-indicators")
//	}
func NewMongoCollection() (*mongo.Client, *mongo.Collection) {
	ctx := context.Background()

	host := os.Getenv("MONGO_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("MONGO_PORT")
	if port == "" {
		port = "27017"
	}

	uri := "mongodb://" + host + ":" + port
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("MongoDB connect error: %v", err)
	}

	return client, client.Database("falconfeeds").Collection("normalized-indicators")
}

// package shared

// import (
// 	"context"
// 	"log"
// 	"time"

// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// func NewMongoCollection() (*mongo.Client, *mongo.Collection) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	client, err := mongo.Connect(ctx, options.Client().
// 		ApplyURI("mongodb://localhost:27017").
// 		SetServerSelectionTimeout(5*time.Second))
// 	if err != nil {
// 		log.Fatalf("Failed to connect to MongoDB: %v", err)
// 	}

// 	// Verify connection
// 	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancelPing()
// 	if err = client.Ping(ctxPing, nil); err != nil {
// 		log.Fatalf("MongoDB ping failed: %v", err)
// 	}

// 	collection := client.Database("test").Collection("stix")
// 	return client, collection
// }
