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

	"github.com/wbach/cs-analyser/apps/api/internal/auth"
	"github.com/wbach/cs-analyser/apps/api/internal/config"
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

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.Region))
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(awsCfg)
	uploader := manager.NewUploader(s3Client)
	sqsClient := sqs.NewFromConfig(awsCfg)

	uploadHandler := upload.NewHandler(uploader, sqsClient, cfg.Bucket, cfg.QueueURL)
	vetoManager := veto.NewManager()

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/admin/whoami", whoamiHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/upload", uploadHandler)
	mux.HandleFunc("/auth/login", auth.NewLoginHandler(cfg.JWTSecret, cfg.AdminUsername, cfg.AdminPasswordHash, adminTokenTTL))
	mux.HandleFunc("/veto/sessions", veto.NewCreateSessionHandler(vetoManager, cfg.JWTSecret, vetoTokenTTL))
	mux.HandleFunc("/veto/ws", veto.NewWebSocketHandler(vetoManager, cfg.JWTSecret, cfg.AllowedOrigin))
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
	if err := json.NewEncoder(w).Encode(map[string]string{"username": claims.Subject}); err != nil {
		log.Printf("failed to write whoami response: %v", err)
	}
}
