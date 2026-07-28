package demos

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const maxDemos = 200

type Demo struct {
	ID          string    `json:"id"`
	MapName     string    `json:"mapName"`
	Winner      string    `json:"winner"`
	TeamTScore  int       `json:"teamTScore"`
	TeamCTScore int       `json:"teamCtScore"`
	PlayerCount int       `json:"playerCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

func NewListHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := loadDemos(r.Context(), pool)
		if err != nil {
			log.Printf("failed to load demos: %v", err)
			http.Error(w, "failed to load demos", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(list); err != nil {
			log.Printf("failed to write demos response: %v", err)
		}
	}
}

func loadDemos(ctx context.Context, pool *pgxpool.Pool) ([]Demo, error) {
	rows, err := pool.Query(
		ctx,
		`SELECT id, coalesce(map_name, ''), coalesce(winner, ''), team_t_score, team_ct_score,
			jsonb_array_length(players), created_at
		FROM demos
		ORDER BY created_at DESC
		LIMIT $1`,
		maxDemos,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Demo{}
	for rows.Next() {
		var id int64
		var demo Demo
		if err := rows.Scan(&id, &demo.MapName, &demo.Winner, &demo.TeamTScore, &demo.TeamCTScore, &demo.PlayerCount, &demo.CreatedAt); err != nil {
			return nil, err
		}
		demo.ID = strconv.FormatInt(id, 10)
		list = append(list, demo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
