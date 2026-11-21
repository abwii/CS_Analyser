package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func newRequestFromIP(ip string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.RemoteAddr = ip + ":12345"
	return req
}

func TestRateLimiterAllowsWithinBurst(t *testing.T) {
	limiter := NewRateLimiter(1, 3)
	handler := limiter.Middleware(okTestHandler())

	for i := 0; i < 3; i++ {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, newRequestFromIP("203.0.113.1"))
		if recorder.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, recorder.Code)
		}
	}
}

func TestRateLimiterRejectsBeyondBurst(t *testing.T) {
	limiter := NewRateLimiter(1, 3)
	handler := limiter.Middleware(okTestHandler())

	for i := 0; i < 3; i++ {
		handler.ServeHTTP(httptest.NewRecorder(), newRequestFromIP("203.0.113.2"))
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, newRequestFromIP("203.0.113.2"))

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", recorder.Code)
	}
}

func TestRateLimiterTracksClientsIndependently(t *testing.T) {
	limiter := NewRateLimiter(1, 1)
	handler := limiter.Middleware(okTestHandler())

	recorderA := httptest.NewRecorder()
	handler.ServeHTTP(recorderA, newRequestFromIP("203.0.113.3"))
	if recorderA.Code != http.StatusOK {
		t.Fatalf("client A: expected 200, got %d", recorderA.Code)
	}

	recorderB := httptest.NewRecorder()
	handler.ServeHTTP(recorderB, newRequestFromIP("203.0.113.4"))
	if recorderB.Code != http.StatusOK {
		t.Fatalf("client B (different IP): expected 200, got %d", recorderB.Code)
	}

	recorderARetry := httptest.NewRecorder()
	handler.ServeHTTP(recorderARetry, newRequestFromIP("203.0.113.3"))
	if recorderARetry.Code != http.StatusTooManyRequests {
		t.Fatalf("client A retry: expected 429 (own burst exhausted), got %d", recorderARetry.Code)
	}
}

func TestClientIPStripsPort(t *testing.T) {
	req := newRequestFromIP("203.0.113.5")
	if got := clientIP(req); got != "203.0.113.5" {
		t.Fatalf("expected 203.0.113.5, got %s", got)
	}
}

func TestClientIPFallsBackToRawRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "not-a-host-port"
	if got := clientIP(req); got != "not-a-host-port" {
		t.Fatalf("expected raw remote addr fallback, got %s", got)
	}
}
