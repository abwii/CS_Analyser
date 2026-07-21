package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func NewLoginHandler(secret []byte, adminUsername string, adminPasswordHash string, ttl time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Username != adminUsername || !CheckPassword(adminPasswordHash, req.Password) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		expiresAt := time.Now().Add(ttl)
		token, err := NewAdminToken(secret, req.Username, ttl)
		if err != nil {
			log.Printf("failed to sign token: %v", err)
			http.Error(w, "failed to issue token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(LoginResponse{Token: token, ExpiresAt: expiresAt}); err != nil {
			log.Printf("failed to write login response: %v", err)
		}
	}
}
