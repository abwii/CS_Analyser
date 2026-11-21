package players

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/wbach/cs-analyser/apps/api/internal/cache"
	"github.com/wbach/cs-analyser/apps/api/internal/faceit"
)

const faceitRecentMatchesLimit = 10

type RecentMatch struct {
	MatchID string `json:"matchId"`
	Map     string `json:"map"`
	Won     bool   `json:"won"`
}

type RecentMatches struct {
	Available      bool          `json:"available"`
	FaceitNickname string        `json:"faceitNickname,omitempty"`
	Matches        []RecentMatch `json:"matches"`
}

// NewRecentMatchesHandler lists a player's recent FACEIT matches so the UI
// can offer manual demo upload per match. FACEIT's demo_url is not surfaced
// here: it isn't directly downloadable (it has to go through FACEIT's
// separate, access-gated Downloads API - see #40), so there is currently no
// automatic-download path to gate on it.
func NewRecentMatchesHandler(faceitClient *faceit.Client, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steamID := r.PathValue("id")
		if _, err := strconv.ParseInt(steamID, 10, 64); err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}

		if !faceitClient.Configured() {
			writeJSON(w, RecentMatches{Available: false, Matches: []RecentMatch{}})
			return
		}

		ctx := r.Context()
		key := "faceit-matches:" + steamID

		if cached, hit, err := cache.GetJSON[RecentMatches](ctx, redisClient, key); err != nil {
			log.Printf("failed to read faceit matches cache: %v", err)
		} else if hit {
			writeJSON(w, cached)
			return
		}

		matches, err := loadRecentMatches(ctx, faceitClient, steamID)
		if err != nil {
			log.Printf("failed to load faceit matches: %v", err)
			writeJSON(w, RecentMatches{Available: false, Matches: []RecentMatch{}})
			return
		}

		if err := cache.SetJSON(ctx, redisClient, key, matches, faceitCacheTTL); err != nil {
			log.Printf("failed to write faceit matches cache: %v", err)
		}

		writeJSON(w, matches)
	}
}

func loadRecentMatches(ctx context.Context, client *faceit.Client, steamID string) (RecentMatches, error) {
	player, err := client.FindPlayerBySteamID(ctx, steamID)
	if err != nil {
		if errors.Is(err, faceit.ErrNotFound) {
			return RecentMatches{Available: false, Matches: []RecentMatch{}}, nil
		}
		return RecentMatches{}, err
	}

	history, err := client.MatchHistory(ctx, player.PlayerID, faceitRecentMatchesLimit)
	if err != nil {
		return RecentMatches{}, err
	}

	matches := make([]RecentMatch, 0, len(history))
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

		matches = append(matches, RecentMatch{
			MatchID: item.MatchID,
			Map:     mapName,
			Won:     faceit.DidPlayerWin(item, player.PlayerID),
		})
	}

	return RecentMatches{
		Available:      true,
		FaceitNickname: player.Nickname,
		Matches:        matches,
	}, nil
}
