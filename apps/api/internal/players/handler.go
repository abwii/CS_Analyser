package players

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/wbach/cs-analyser/apps/api/internal/cache"
)

const maxPositions = 2000
const minBoundsPadding = 50.0
const boundsPaddingRatio = 0.05
const profileCacheTTL = 60 * time.Second

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
	Z float64 `json:"z"`
}

type Profile struct {
	PlayerID   string  `json:"playerId"`
	PlayerName string  `json:"playerName"`
	Metrics    Metrics `json:"metrics"`
	Map        Map     `json:"map"`
	Positions  []Point `json:"positions"`
}

func NewProfileHandler(pool *pgxpool.Pool, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steamID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		key := profileCacheKey(steamID)

		if cached, hit, err := cache.GetJSON[Profile](ctx, redisClient, key); err != nil {
			log.Printf("failed to read player profile cache: %v", err)
		} else if hit {
			writeJSON(w, cached)
			return
		}

		profile, found, err := loadProfile(ctx, pool, steamID)
		if err != nil {
			log.Printf("failed to load player profile: %v", err)
			http.Error(w, "failed to load player profile", http.StatusInternalServerError)
			return
		}
		if !found {
			http.Error(w, "player not found", http.StatusNotFound)
			return
		}

		if err := cache.SetJSON(ctx, redisClient, key, profile, profileCacheTTL); err != nil {
			log.Printf("failed to write player profile cache: %v", err)
		}

		writeJSON(w, profile)
	}
}

func profileCacheKey(steamID int64) string {
	return fmt.Sprintf("player-profile:%d", steamID)
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write player profile response: %v", err)
	}
}

func loadProfile(ctx context.Context, pool *pgxpool.Pool, steamID int64) (Profile, bool, error) {
	var total int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM player_stats WHERE steamid = $1", steamID).Scan(&total); err != nil {
		return Profile{}, false, err
	}
	if total == 0 {
		return Profile{}, false, nil
	}

	var metrics Metrics
	row := pool.QueryRow(
		ctx,
		"SELECT avg(kpr), avg(dpr), avg(kast), avg(impact), avg(adr) FROM player_stats WHERE steamid = $1",
		steamID,
	)
	if err := row.Scan(&metrics.KPR, &metrics.DPR, &metrics.KAST, &metrics.Impact, &metrics.ADR); err != nil {
		return Profile{}, false, err
	}

	positions, err := loadPositions(ctx, pool, steamID)
	if err != nil {
		return Profile{}, false, err
	}

	playerName, mapName := findPlayerNameAndMap(ctx, pool, steamID)

	profile := Profile{
		PlayerID:   strconv.FormatInt(steamID, 10),
		PlayerName: playerName,
		Metrics:    metrics,
		Map:        Map{Name: mapName, Bounds: computeBounds(positions)},
		Positions:  positions,
	}
	return profile, true, nil
}

func loadPositions(ctx context.Context, pool *pgxpool.Pool, steamID int64) ([]Point, error) {
	rows, err := pool.Query(
		ctx,
		"SELECT x, y, z FROM player_positions WHERE steamid = $1 ORDER BY id DESC LIMIT $2",
		steamID,
		maxPositions,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []Point{}
	for rows.Next() {
		var point Point
		if err := rows.Scan(&point.X, &point.Y, &point.Z); err != nil {
			return nil, err
		}
		positions = append(positions, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func computeBounds(positions []Point) WorldBounds {
	if len(positions) == 0 {
		return WorldBounds{}
	}

	minX, maxX := positions[0].X, positions[0].X
	minY, maxY := positions[0].Y, positions[0].Y

	for _, point := range positions[1:] {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	padX := boundsPaddingRatio * (maxX - minX)
	if padX < minBoundsPadding {
		padX = minBoundsPadding
	}
	padY := boundsPaddingRatio * (maxY - minY)
	if padY < minBoundsPadding {
		padY = minBoundsPadding
	}

	return WorldBounds{
		MinX: minX - padX,
		MaxX: maxX + padX,
		MinY: minY - padY,
		MaxY: maxY + padY,
	}
}

func findPlayerNameAndMap(ctx context.Context, pool *pgxpool.Pool, steamID int64) (string, string) {
	fallback := fmt.Sprintf("Player %d", steamID)
	filter := fmt.Sprintf(`[{"steamid": %d}]`, steamID)

	var roster []byte
	var mapName *string
	err := pool.QueryRow(
		ctx,
		"SELECT players, map_name FROM demos WHERE players @> $1::jsonb ORDER BY created_at DESC LIMIT 1",
		filter,
	).Scan(&roster, &mapName)
	if err != nil {
		return fallback, ""
	}

	name := fallback
	var entries []struct {
		SteamID uint64 `json:"steamid"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(roster, &entries); err == nil {
		for _, entry := range entries {
			if int64(entry.SteamID) == steamID && entry.Name != "" {
				name = entry.Name
				break
			}
		}
	}

	if mapName == nil {
		return name, ""
	}
	return name, *mapName
}
