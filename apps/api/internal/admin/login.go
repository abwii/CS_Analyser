package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wbach/cs-analyser/apps/api/internal/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func NewLoginHandler(pool *pgxpool.Pool, secret []byte, ttl time.Duration) http.HandlerFunc {
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

		var passwordHash string
		err := pool.QueryRow(
			r.Context(),
			"SELECT coalesce(password_hash, '') FROM admin_users WHERE email = $1",
			req.Email,
		).Scan(&passwordHash)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if !auth.CheckPassword(passwordHash, req.Password) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		expiresAt := time.Now().Add(ttl)
		token, err := auth.NewAdminToken(secret, req.Email, ttl)
		if err != nil {
			log.Printf("failed to sign token: %v", err)
			http.Error(w, "failed to issue token", http.StatusInternalServerError)
			return
		}

		writeJSON(w, LoginResponse{Token: token, ExpiresAt: expiresAt})
	}
}
