package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !CheckPassword(hash, "correct-password") {
		t.Fatalf("expected matching password to succeed")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatalf("expected wrong password to fail")
	}
}

func TestCheckPasswordRejectsMalformedHash(t *testing.T) {
	if CheckPassword("not-a-bcrypt-hash", "anything") {
		t.Fatalf("expected malformed hash to fail")
	}
}
