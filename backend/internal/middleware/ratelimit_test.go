package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimit_AllowsBurstThenBlocks(t *testing.T) {
	handler := RateLimit(1, 2)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "203.0.113.1:12345"
		return req
	}

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, newReq())
		require.Equal(t, http.StatusOK, rec.Code, "request %d within burst should succeed", i)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newReq())
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimit_TracksClientsIndependently(t *testing.T) {
	handler := RateLimit(1, 1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	reqA := httptest.NewRequest(http.MethodGet, "/health", nil)
	reqA.RemoteAddr = "203.0.113.1:1"
	recA := httptest.NewRecorder()
	handler.ServeHTTP(recA, reqA)
	require.Equal(t, http.StatusOK, recA.Code)

	reqB := httptest.NewRequest(http.MethodGet, "/health", nil)
	reqB.RemoteAddr = "203.0.113.2:1"
	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, reqB)
	require.Equal(t, http.StatusOK, recB.Code)
}
