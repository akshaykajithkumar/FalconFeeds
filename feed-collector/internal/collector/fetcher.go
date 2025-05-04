package collector

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var feeds = []string{
	//"https://bazaar.abuse.ch/export/txt/sha256/full/",
	//"https://www.cert-bund.de/feeds/advisoryde.xml",
	"https://data.phishtank.com/data/online-valid.json",
}

func FetchAndPublishFeeds(redisClient *redis.Client, tracer trace.Tracer, metric *prometheus.CounterVec) {
	ctx := context.Background()

	for _, url := range feeds {
		func() {
			ctx, span := tracer.Start(ctx, "fetch-feed")
			defer span.End()
			span.SetAttributes(attribute.String("url", url))

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Error fetching %s: %v", url, err)
				metric.WithLabelValues("failure").Inc()
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response from %s: %v", url, err)
				metric.WithLabelValues("failure").Inc()
				return
			}

			_, err = redisClient.XAdd(ctx, &redis.XAddArgs{
				Stream: "raw-feeds",
				Values: map[string]interface{}{
					"url":     url,
					"payload": string(body),
				},
			}).Result()

			if err != nil {
				log.Printf("Error publishing to Redis: %v", err)
				metric.WithLabelValues("failure").Inc()
				return
			}

			log.Printf("Published feed from %s", url)
			metric.WithLabelValues("success").Inc()
		}()
	}
}
