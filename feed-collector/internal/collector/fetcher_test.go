package collector

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestFetchAndPublishFeeds(t *testing.T) {
	t.Run("successful publish", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("test-data"))
		}))
		defer ts.Close()

		rdb, mock := redismock.NewClientMock()
		mock.ExpectXAdd(&redis.XAddArgs{
			Stream: "raw-feeds",
			Values: map[string]interface{}{
				"url":     ts.URL,
				"payload": "test-data",
			},
		}).SetVal("1-0")

		originalFeeds := feeds
		feeds = []string{ts.URL}
		defer func() { feeds = originalFeeds }()

		tracer := trace.NewNoopTracerProvider().Tracer("test")
		metric := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test"}, []string{"status"})

		FetchAndPublishFeeds(rdb, tracer, metric)

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, 1.0, testutil.ToFloat64(metric.WithLabelValues("success")))
	})

	t.Run("http failure", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		tracer := trace.NewNoopTracerProvider().Tracer("test")
		metric := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test"}, []string{"status"})

		// Use URL format that fails parsing without DNS lookup
		originalFeeds := feeds
		feeds = []string{"://invalid-url"}
		defer func() { feeds = originalFeeds }()

		FetchAndPublishFeeds(rdb, tracer, metric)

		// Should have no Redis interactions
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, 1.0, testutil.ToFloat64(metric.WithLabelValues("failure")))
	})

	t.Run("redis failure", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("test-data"))
		}))
		defer ts.Close()

		rdb, mock := redismock.NewClientMock()
		mock.ExpectXAdd(&redis.XAddArgs{
			Stream: "raw-feeds",
			Values: map[string]interface{}{
				"url":     ts.URL,
				"payload": "test-data",
			},
		}).SetErr(errors.New("redis error"))

		tracer := trace.NewNoopTracerProvider().Tracer("test")
		metric := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test"}, []string{"status"})

		originalFeeds := feeds
		feeds = []string{ts.URL}
		defer func() { feeds = originalFeeds }()

		FetchAndPublishFeeds(rdb, tracer, metric)

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, 1.0, testutil.ToFloat64(metric.WithLabelValues("failure")))
	})
}
