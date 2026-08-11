package steamauth

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/wbach/cs-analyser/apps/api/internal/cache"
)

const summaryCacheTTL = time.Hour

type SteamSummary struct {
	Available   bool   `json:"available"`
	PersonaName string `json:"personaName,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
}

func NewSummaryHandler(client *SummaryClient, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steamID := r.PathValue("id")
		if _, err := strconv.ParseInt(steamID, 10, 64); err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}

		if !client.Configured() {
			writeJSON(w, SteamSummary{Available: false})
			return
		}

		ctx := r.Context()
		key := "steam-summary:" + steamID

		if cached, hit, err := cache.GetJSON[SteamSummary](ctx, redisClient, key); err != nil {
			log.Printf("failed to read steam summary cache: %v", err)
		} else if hit {
			writeJSON(w, cached)
			return
		}

		summary, err := client.GetPlayerSummary(ctx, steamID)
		if err != nil {
			log.Printf("failed to load steam summary: %v", err)
			writeJSON(w, SteamSummary{Available: false})
			return
		}

		result := SteamSummary{Available: true, PersonaName: summary.PersonaName, AvatarURL: summary.AvatarURL}

		if err := cache.SetJSON(ctx, redisClient, key, result, summaryCacheTTL); err != nil {
			log.Printf("failed to write steam summary cache: %v", err)
		}

		writeJSON(w, result)
	}
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write steam summary response: %v", err)
	}
}
