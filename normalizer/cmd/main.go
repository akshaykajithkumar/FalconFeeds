package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"normalizer/internal/handler"
	"normalizer/internal/normalizer"
	"normalizer/internal/shared"
)

func main() {
	redisClient := shared.NewRedisClient()
	mongoClient, mongoColl := shared.NewMongoCollection()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	processor := normalizer.NewProcessor(redisClient, mongoColl)
	go processor.Start(ctx)

	apiHandler := handler.NewAPIHandler(mongoColl)
	router := apiHandler.RegisterRoutes()

	srv := &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	go func() {
		log.Println("Starting API server on :5000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down...")

	cancel()
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		log.Printf("MongoDB disconnect error: %v", err)
	}

	redisClient.Close()
}
