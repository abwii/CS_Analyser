package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/wbach/cs-analyser/apps/api/internal/config"
	"github.com/wbach/cs-analyser/apps/api/internal/upload"
	"github.com/wbach/cs-analyser/apps/api/internal/veto"
)

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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/upload", uploadHandler)
	mux.HandleFunc("/veto/sessions", veto.NewCreateSessionHandler(vetoManager))
	mux.HandleFunc("/veto/ws", veto.NewWebSocketHandler(vetoManager))

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
