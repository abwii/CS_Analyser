package admin

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const maxPageSize = 100

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type PaginatedUsers struct {
	Users    []User `json:"users"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func NewUsersHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := parsePositiveInt(r.URL.Query().Get("page"), 1)
		pageSize := parsePositiveInt(r.URL.Query().Get("pageSize"), 10)
		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}

		ctx := r.Context()

		var total int
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM admin_users").Scan(&total); err != nil {
			log.Printf("failed to count admin users: %v", err)
			http.Error(w, "failed to load users", http.StatusInternalServerError)
			return
		}

		rows, err := pool.Query(
			ctx,
			"SELECT id, email, role, created_at FROM admin_users ORDER BY id ASC LIMIT $1 OFFSET $2",
			pageSize,
			(page-1)*pageSize,
		)
		if err != nil {
			log.Printf("failed to query admin users: %v", err)
			http.Error(w, "failed to load users", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var id int64
			var user User
			if err := rows.Scan(&id, &user.Email, &user.Role, &user.CreatedAt); err != nil {
				log.Printf("failed to scan admin user: %v", err)
				http.Error(w, "failed to load users", http.StatusInternalServerError)
				return
			}
			user.ID = strconv.FormatInt(id, 10)
			users = append(users, user)
		}
		if err := rows.Err(); err != nil {
			log.Printf("failed to iterate admin users: %v", err)
			http.Error(w, "failed to load users", http.StatusInternalServerError)
			return
		}

		writeJSON(w, PaginatedUsers{Users: users, Total: total, Page: page, PageSize: pageSize})
	}
}
