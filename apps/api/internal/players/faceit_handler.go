package players

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/wbach/cs-analyser/apps/api/internal/cache"
	"github.com/wbach/cs-analyser/apps/api/internal/faceit"
)

const faceitCacheTTL = 10 * time.Minute
const faceitMatchLimit = 20

type FaceitStats struct {
	Available      bool              `json:"available"`
	FaceitNickname string            `json:"faceitNickname,omitempty"`
	ByMap          []faceit.MapStats `json:"byMap"`
}

func NewFaceitStatsHandler(faceitClient *faceit.Client, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steamID := r.PathValue("id")
		if _, err := strconv.ParseInt(steamID, 10, 64); err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}

		if !faceitClient.Configured() {
			writeJSON(w, FaceitStats{Available: false, ByMap: []faceit.MapStats{}})
			return
		}

		ctx := r.Context()
		key := "faceit-stats:" + steamID

		if cached, hit, err := cache.GetJSON[FaceitStats](ctx, redisClient, key); err != nil {
			log.Printf("failed to read faceit stats cache: %v", err)
		} else if hit {
			writeJSON(w, cached)
			return
		}

		stats, err := loadFaceitStats(ctx, faceitClient, steamID)
		if err != nil {
			log.Printf("failed to load faceit stats: %v", err)
			writeJSON(w, FaceitStats{Available: false, ByMap: []faceit.MapStats{}})
			return
		}

		if err := cache.SetJSON(ctx, redisClient, key, stats, faceitCacheTTL); err != nil {
			log.Printf("failed to write faceit stats cache: %v", err)
		}

		writeJSON(w, stats)
	}
}

func loadFaceitStats(ctx context.Context, client *faceit.Client, steamID string) (FaceitStats, error) {
	player, err := client.FindPlayerBySteamID(ctx, steamID)
	if err != nil {
		if errors.Is(err, faceit.ErrNotFound) {
			return FaceitStats{Available: false, ByMap: []faceit.MapStats{}}, nil
		}
		return FaceitStats{}, err
	}

	history, err := client.MatchHistory(ctx, player.PlayerID, faceitMatchLimit)
	if err != nil {
		return FaceitStats{}, err
	}

	results := make([]faceit.MatchResult, 0, len(history))
	for _, item := range history {
		matchStats, err := client.MatchStats(ctx, item.MatchID)
		if err != nil {
			log.Printf("failed to load faceit stats for match %s: %v", item.MatchID, err)
			continue
		}

		mapName := faceit.ExtractMap(matchStats)
		if mapName == "" {
			continue
		}

		result := faceit.MatchResult{
			Map: mapName,
			Won: faceit.DidPlayerWin(item, player.PlayerID),
		}
		if kills, deaths, ok := faceit.ExtractKD(matchStats, player.PlayerID); ok {
			result.Kills = kills
			result.Deaths = deaths
			result.HasKD = true
		}
		results = append(results, result)
	}

	return FaceitStats{
		Available:      true,
		FaceitNickname: player.Nickname,
		ByMap:          faceit.AggregateByMap(results),
	}, nil
}
