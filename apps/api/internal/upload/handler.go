package upload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type S3Uploader interface {
	//nolint:staticcheck // feature/s3/transfermanager replacement is still pre-1.0 (v0.5.x); revisit once stable, see #59
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

type SQSPublisher interface {
	SendMessage(ctx context.Context, input *sqs.SendMessageInput, opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type Handler struct {
	Uploader  S3Uploader
	SQSClient SQSPublisher
	Bucket    string
	QueueURL  string
}

func NewHandler(uploader S3Uploader, sqsClient SQSPublisher, bucket, queueURL string) *Handler {
	return &Handler{
		Uploader:  uploader,
		SQSClient: sqsClient,
		Bucket:    bucket,
		QueueURL:  queueURL,
	}
}

type Message struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "invalid multipart request", http.StatusBadRequest)
		return
	}

	var key string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "failed to read multipart data", http.StatusBadRequest)
			return
		}

		if part.FormName() != "file" {
			part.Close()
			continue
		}

		key, err = generateKey(part.FileName())
		if err != nil {
			part.Close()
			http.Error(w, "failed to generate object key", http.StatusInternalServerError)
			return
		}

		_, err = h.Uploader.Upload(r.Context(), &s3.PutObjectInput{
			Bucket: aws.String(h.Bucket),
			Key:    aws.String(key),
			Body:   part,
		})
		part.Close()
		if err != nil {
			log.Printf("s3 upload failed: %v", err)
			http.Error(w, "failed to upload to s3", http.StatusInternalServerError)
			return
		}

		break
	}

	if key == "" {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}

	msg := Message{
		Bucket: h.Bucket,
		Key:    key,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "failed to encode message", http.StatusInternalServerError)
		return
	}

	_, err = h.SQSClient.SendMessage(r.Context(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(h.QueueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		log.Printf("sqs publish failed: %v", err)
		http.Error(w, "failed to publish sqs message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func generateKey(filename string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	ext := filepath.Ext(filename)
	return fmt.Sprintf("uploads/%s%s", hex.EncodeToString(buf), ext), nil
}
