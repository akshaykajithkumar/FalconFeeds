package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"feed-collector/internal/collector"
	"feed-collector/internal/shared"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
)

var (
	requestsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_processed_total",
			Help: "Total number of processed requests",
		},
		[]string{"method", "status"},
	)

	feedsFetched = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "feeds_fetched_total",
			Help: "Number of feeds fetched, labeled by status",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(requestsProcessed)
	prometheus.MustRegister(feedsFetched)
}

func main() {
	// Init tracing
	tracerProvider := shared.InitTracer()
	defer func() {
		_ = tracerProvider.Shutdown(context.Background())
	}()
	tracer := otel.Tracer("feed-collector")

	// Init Redis
	redisClient := shared.NewRedisClient()

	// Handle signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	// Fetch interval
	interval := 5 * time.Minute
	if env := os.Getenv("FETCH_INTERVAL_MINUTES"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			interval = time.Duration(val) * time.Minute
		}
	}
	ticker := time.NewTicker(interval)

	// Periodic fetching
	go func() {
		for {
			select {
			case <-ticker.C:
				collector.FetchAndPublishFeeds(redisClient, tracer, feedsFetched)
			case <-ctx.Done():
				log.Println("Ticker stopped")
				return
			}
		}
	}()

	// Immediate fetch on start
	collector.FetchAndPublishFeeds(redisClient, tracer, feedsFetched)

	// HTTP server setup
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		requestsProcessed.WithLabelValues(r.Method, "200").Inc()
		fmt.Fprintln(w, "OK")
	})

	// Prometheus
	mux.Handle("/metrics", promhttp.Handler())

	// Start HTTP server
	go func() {
		log.Println("HTTP server listening on :4000")
		if err := http.ListenAndServe(":4000", mux); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for shutdown
	<-quit
	log.Println("Shutting down Feed Collector...")
	cancel()
	ticker.Stop()
}
