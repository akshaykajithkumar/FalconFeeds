package shared

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
