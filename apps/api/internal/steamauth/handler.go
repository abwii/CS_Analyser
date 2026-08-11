package steamauth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/wbach/cs-analyser/apps/api/internal/auth"
)

func NewLoginHandler(returnTo, realm string) http.HandlerFunc {
	loginURL := BuildLoginURL(returnTo, realm)
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, loginURL, http.StatusFound)
	}
}

func NewCallbackHandler(verifier *Verifier, jwtSecret []byte, ttl time.Duration, webBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steamID, err := verifier.Verify(r.Context(), r.URL.Query())
		if err != nil {
			log.Printf("steam openid verification failed: %v", err)
			http.Redirect(w, r, webBaseURL+"/login?error=steam", http.StatusFound)
			return
		}

		token, err := auth.NewSteamToken(jwtSecret, steamID, ttl)
		if err != nil {
			log.Printf("failed to issue steam token: %v", err)
			http.Redirect(w, r, webBaseURL+"/login?error=steam", http.StatusFound)
			return
		}

		redirectURL := fmt.Sprintf("%s/auth/steam/callback?token=%s", webBaseURL, url.QueryEscape(token))
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}
