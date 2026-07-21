package admin

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompetitiveMap struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type mapInput struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func NewListMapsHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), "SELECT id, name, active FROM competitive_maps ORDER BY name ASC")
		if err != nil {
			log.Printf("failed to query maps: %v", err)
			http.Error(w, "failed to load maps", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		maps := []CompetitiveMap{}
		for rows.Next() {
			var id int64
			var m CompetitiveMap
			if err := rows.Scan(&id, &m.Name, &m.Active); err != nil {
				log.Printf("failed to scan map: %v", err)
				http.Error(w, "failed to load maps", http.StatusInternalServerError)
				return
			}
			m.ID = strconv.FormatInt(id, 10)
			maps = append(maps, m)
		}
		if err := rows.Err(); err != nil {
			log.Printf("failed to iterate maps: %v", err)
			http.Error(w, "failed to load maps", http.StatusInternalServerError)
			return
		}

		writeJSON(w, maps)
	}
}

func NewCreateMapHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input mapInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if input.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		var id int64
		row := pool.QueryRow(
			r.Context(),
			"INSERT INTO competitive_maps (name, active) VALUES ($1, $2) RETURNING id",
			input.Name,
			input.Active,
		)
		if err := row.Scan(&id); err != nil {
			log.Printf("failed to create map: %v", err)
			http.Error(w, "failed to create map", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		writeJSON(w, CompetitiveMap{ID: strconv.FormatInt(id, 10), Name: input.Name, Active: input.Active})
	}
}

func NewUpdateMapHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid map id", http.StatusBadRequest)
			return
		}

		var input mapInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if input.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		row := pool.QueryRow(
			r.Context(),
			"UPDATE competitive_maps SET name = $1, active = $2 WHERE id = $3 RETURNING id",
			input.Name,
			input.Active,
			id,
		)
		var updatedID int64
		if err := row.Scan(&updatedID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "map not found", http.StatusNotFound)
				return
			}
			log.Printf("failed to update map: %v", err)
			http.Error(w, "failed to update map", http.StatusInternalServerError)
			return
		}

		writeJSON(w, CompetitiveMap{ID: strconv.FormatInt(updatedID, 10), Name: input.Name, Active: input.Active})
	}
}

func NewDeleteMapHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid map id", http.StatusBadRequest)
			return
		}

		tag, err := pool.Exec(r.Context(), "DELETE FROM competitive_maps WHERE id = $1", id)
		if err != nil {
			log.Printf("failed to delete map: %v", err)
			http.Error(w, "failed to delete map", http.StatusInternalServerError)
			return
		}
		if tag.RowsAffected() == 0 {
			http.Error(w, "map not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
