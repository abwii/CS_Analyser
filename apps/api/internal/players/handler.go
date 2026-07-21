package players

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Metrics struct {
	KPR    float64 `json:"kpr"`
	DPR    float64 `json:"dpr"`
	KAST   float64 `json:"kast"`
	Impact float64 `json:"impact"`
	ADR    float64 `json:"adr"`
}

type WorldBounds struct {
	MinX float64 `json:"minX"`
	MaxX float64 `json:"maxX"`
	MinY float64 `json:"minY"`
	MaxY float64 `json:"maxY"`
}

type Map struct {
	Name     string      `json:"name"`
	ImageURL string      `json:"imageUrl"`
	Bounds   WorldBounds `json:"bounds"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Profile struct {
	PlayerID   string  `json:"playerId"`
	PlayerName string  `json:"playerName"`
	Metrics    Metrics `json:"metrics"`
	Map        Map     `json:"map"`
	Positions  []Point `json:"positions"`
}

func NewProfileHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steamID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		var total int
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM player_stats WHERE steamid = $1", steamID).Scan(&total); err != nil {
			log.Printf("failed to count player stats: %v", err)
			http.Error(w, "failed to load player profile", http.StatusInternalServerError)
			return
		}
		if total == 0 {
			http.Error(w, "player not found", http.StatusNotFound)
			return
		}

		var metrics Metrics
		row := pool.QueryRow(
			ctx,
			"SELECT avg(kpr), avg(dpr), avg(kast), avg(impact), avg(adr) FROM player_stats WHERE steamid = $1",
			steamID,
		)
		if err := row.Scan(&metrics.KPR, &metrics.DPR, &metrics.KAST, &metrics.Impact, &metrics.ADR); err != nil {
			log.Printf("failed to load player metrics: %v", err)
			http.Error(w, "failed to load player profile", http.StatusInternalServerError)
			return
		}

		profile := Profile{
			PlayerID:   strconv.FormatInt(steamID, 10),
			PlayerName: findPlayerName(ctx, pool, steamID),
			Metrics:    metrics,
			Map:        Map{Bounds: WorldBounds{}},
			Positions:  []Point{},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			log.Printf("failed to write player profile response: %v", err)
		}
	}
}

func findPlayerName(ctx context.Context, pool *pgxpool.Pool, steamID int64) string {
	fallback := fmt.Sprintf("Player %d", steamID)
	filter := fmt.Sprintf(`[{"steamid": %d}]`, steamID)

	var raw []byte
	err := pool.QueryRow(
		ctx,
		"SELECT players FROM demos WHERE players @> $1::jsonb ORDER BY created_at DESC LIMIT 1",
		filter,
	).Scan(&raw)
	if err != nil {
		return fallback
	}

	var roster []struct {
		SteamID uint64 `json:"steamid"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(raw, &roster); err != nil {
		return fallback
	}

	for _, entry := range roster {
		if int64(entry.SteamID) == steamID && entry.Name != "" {
			return entry.Name
		}
	}
	return fallback
}
