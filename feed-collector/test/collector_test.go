package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeedFetchMock(t *testing.T) {
	mockFeed := "<rss><channel><title>Test Feed</title></channel></rss>"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockFeed))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Test Feed")
}
