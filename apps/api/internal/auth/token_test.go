package auth

import (
	"testing"
	"time"
)

var testSecret = []byte("test-secret-value-that-is-long-enough-for-hs256")

func TestNewAdminTokenRoundTrips(t *testing.T) {
	token, err := NewAdminToken(testSecret, "alice", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := Parse(testSecret, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Scope != ScopeAdmin {
		t.Fatalf("expected scope admin, got %s", claims.Scope)
	}
	if claims.Subject != "alice" {
		t.Fatalf("expected subject alice, got %s", claims.Subject)
	}
}

func TestNewVetoTokenRoundTrips(t *testing.T) {
	token, err := NewVetoToken(testSecret, "session-1", "TeamA", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := Parse(testSecret, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Scope != ScopeVeto {
		t.Fatalf("expected scope veto, got %s", claims.Scope)
	}
	if claims.SessionID != "session-1" {
		t.Fatalf("expected session-1, got %s", claims.SessionID)
	}
	if claims.Team != "TeamA" {
		t.Fatalf("expected TeamA, got %s", claims.Team)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	token, err := NewAdminToken(testSecret, "alice", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wrongSecret := []byte("a-completely-different-secret-value-of-similar-length")
	if _, err := Parse(wrongSecret, token); err == nil {
		t.Fatalf("expected parse to fail with the wrong secret")
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	token, err := NewAdminToken(testSecret, "alice", -time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := Parse(testSecret, token); err == nil {
		t.Fatalf("expected parse to fail for an expired token")
	}
}

func TestParseRejectsMalformedToken(t *testing.T) {
	if _, err := Parse(testSecret, "not-a-valid-jwt"); err == nil {
		t.Fatalf("expected parse to fail for a malformed token")
	}
}

func TestParseRejectsUnsignedAlgNone(t *testing.T) {
	unsignedToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		"eyJzY29wZSI6ImFkbWluIiwic3ViIjoiYWxpY2UifQ."
	if _, err := Parse(testSecret, unsignedToken); err == nil {
		t.Fatalf("expected parse to reject an alg=none token")
	}
}
