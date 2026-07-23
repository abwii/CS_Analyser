package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/wbach/cs-analyser/apps/api/internal/admin"
	"github.com/wbach/cs-analyser/apps/api/internal/auth"
	"github.com/wbach/cs-analyser/apps/api/internal/config"
	"github.com/wbach/cs-analyser/apps/api/internal/db"
	"github.com/wbach/cs-analyser/apps/api/internal/players"
	"github.com/wbach/cs-analyser/apps/api/internal/upload"
	"github.com/wbach/cs-analyser/apps/api/internal/veto"
)

const adminTokenTTL = 8 * time.Hour
const vetoTokenTTL = 4 * time.Hour

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.S3ForcePathStyle
	})
	uploader := manager.NewUploader(s3Client)
	sqsClient := sqs.NewFromConfig(awsCfg)

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	uploadHandler := upload.NewHandler(uploader, sqsClient, cfg.Bucket, cfg.QueueURL)
	vetoManager := veto.NewManager()

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/admin/whoami", whoamiHandler)
	adminMux.HandleFunc("GET /admin/users", admin.NewUsersHandler(pool))
	adminMux.HandleFunc("GET /admin/queue-size", admin.NewQueueSizeHandler(sqsClient, cfg.QueueURL))
	adminMux.HandleFunc("GET /admin/maps", admin.NewListMapsHandler(pool))
	adminMux.HandleFunc("POST /admin/maps", admin.NewCreateMapHandler(pool))
	adminMux.HandleFunc("PUT /admin/maps/{id}", admin.NewUpdateMapHandler(pool))
	adminMux.HandleFunc("DELETE /admin/maps/{id}", admin.NewDeleteMapHandler(pool))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/upload", uploadHandler)
	mux.HandleFunc("/auth/login", admin.NewLoginHandler(pool, cfg.JWTSecret, adminTokenTTL))
	mux.Handle("/veto/sessions", auth.RequireAuth(cfg.JWTSecret, auth.ScopeAdmin)(veto.NewCreateSessionHandler(vetoManager, cfg.JWTSecret, vetoTokenTTL)))
	mux.HandleFunc("/veto/ws", veto.NewWebSocketHandler(vetoManager, cfg.JWTSecret, cfg.AllowedOrigin))
	mux.HandleFunc("GET /players/{id}/profile", players.NewProfileHandler(pool))
	mux.Handle("/admin/", auth.RequireAuth(cfg.JWTSecret, auth.ScopeAdmin)(adminMux))

	handler := auth.CORS(cfg.AllowedOrigin)(mux)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func whoamiHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"email": claims.Subject}); err != nil {
		log.Printf("failed to write whoami response: %v", err)
	}
}
