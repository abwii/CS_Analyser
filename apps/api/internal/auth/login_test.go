package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginHandlerIssuesTokenForValidCredentials(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := NewLoginHandler(testSecret, "admin", hash, time.Hour)
	body, err := json.Marshal(LoginRequest{Username: "admin", Password: "correct-password"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	claims, err := Parse(testSecret, loginResp.Token)
	if err != nil {
		t.Fatalf("expected issued token to be valid: %v", err)
	}
	if claims.Subject != "admin" {
		t.Fatalf("expected subject admin, got %s", claims.Subject)
	}
}

func TestLoginHandlerRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := NewLoginHandler(testSecret, "admin", hash, time.Hour)
	body, err := json.Marshal(LoginRequest{Username: "admin", Password: "wrong-password"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestLoginHandlerRejectsUnknownUsername(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := NewLoginHandler(testSecret, "admin", hash, time.Hour)
	body, err := json.Marshal(LoginRequest{Username: "someone-else", Password: "correct-password"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	handler(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestLoginHandlerRejectsNonPost(t *testing.T) {
	handler := NewLoginHandler(testSecret, "admin", "irrelevant", time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	recorder := httptest.NewRecorder()

	handler(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}
}
