package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSSetsHeadersForAllowedOrigin(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS([]string{"https://example.com", "https://other.example.com"})(inner)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://example.com")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if !called {
		t.Fatalf("expected inner handler to be called")
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("unexpected allow origin header: %s", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("unexpected allow credentials header: %s", got)
	}
}

func TestCORSAllowsSecondConfiguredOrigin(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS([]string{"https://example.com", "https://other.example.com"})(inner)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://other.example.com")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://other.example.com" {
		t.Fatalf("unexpected allow origin header: %s", got)
	}
}

func TestCORSOmitsHeaderForDisallowedOrigin(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS([]string{"https://example.com"})(inner)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://attacker.example")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if !called {
		t.Fatalf("expected inner handler to still be called")
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow origin header, got %s", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected no allow credentials header, got %s", got)
	}
}

func TestCORSShortCircuitsPreflight(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := CORS([]string{"https://example.com"})(inner)
	req := httptest.NewRequest(http.MethodOptions, "/admin/whoami", nil)
	req.Header.Set("Origin", "https://example.com")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if called {
		t.Fatalf("expected inner handler not to be called for preflight")
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}
