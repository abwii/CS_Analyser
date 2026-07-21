package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func protectedTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "missing claims", http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-Subject", claims.Subject)
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuthAcceptsValidToken(t *testing.T) {
	token, err := NewAdminToken(testSecret, "alice", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := RequireAuth(testSecret, ScopeAdmin)(protectedTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/admin/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Subject") != "alice" {
		t.Fatalf("expected subject alice, got %s", recorder.Header().Get("X-Subject"))
	}
}

func TestRequireAuthRejectsMissingHeader(t *testing.T) {
	handler := RequireAuth(testSecret, ScopeAdmin)(protectedTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/admin/whoami", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestRequireAuthRejectsInvalidToken(t *testing.T) {
	handler := RequireAuth(testSecret, ScopeAdmin)(protectedTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/admin/whoami", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestRequireAuthRejectsWrongScope(t *testing.T) {
	token, err := NewVetoToken(testSecret, "session-1", "A", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := RequireAuth(testSecret, ScopeAdmin)(protectedTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/admin/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestRequireAuthRejectsMalformedAuthorizationHeader(t *testing.T) {
	token, err := NewAdminToken(testSecret, "alice", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := RequireAuth(testSecret, ScopeAdmin)(protectedTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/admin/whoami", nil)
	req.Header.Set("Authorization", token)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}
