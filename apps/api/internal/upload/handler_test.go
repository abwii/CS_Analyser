package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type fakeUploader struct {
	err      error
	uploaded []byte
}

//nolint:staticcheck // feature/s3/transfermanager replacement is still pre-1.0 (v0.5.x); revisit once stable, see #59
func (f *fakeUploader) Upload(_ context.Context, input *s3.PutObjectInput, _ ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	buf, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	f.uploaded = buf
	return &manager.UploadOutput{}, nil
}

type fakeSQSPublisher struct {
	err  error
	sent *sqs.SendMessageInput
}

func (f *fakeSQSPublisher) SendMessage(_ context.Context, input *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.sent = input
	return &sqs.SendMessageOutput{}, nil
}

func TestPublishForParsingUploadsAndPublishes(t *testing.T) {
	uploader := &fakeUploader{}
	sqsClient := &fakeSQSPublisher{}
	handler := NewHandler(uploader, sqsClient, "test-bucket", "test-queue")

	msg, err := handler.PublishForParsing(context.Background(), "uploads/abc.dem", bytes.NewReader([]byte("demo bytes")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Bucket != "test-bucket" || msg.Key != "uploads/abc.dem" {
		t.Fatalf("unexpected message: %+v", msg)
	}
	if string(uploader.uploaded) != "demo bytes" {
		t.Fatalf("unexpected uploaded body: %s", uploader.uploaded)
	}
	if sqsClient.sent == nil {
		t.Fatal("expected an sqs message to be sent")
	}

	var sentMsg Message
	if err := json.Unmarshal([]byte(*sqsClient.sent.MessageBody), &sentMsg); err != nil {
		t.Fatalf("failed to decode sent sqs body: %v", err)
	}
	if sentMsg != msg {
		t.Fatalf("sqs message body %+v does not match returned message %+v", sentMsg, msg)
	}
}

func TestPublishForParsingReturnsErrorOnS3Failure(t *testing.T) {
	uploader := &fakeUploader{err: errors.New("s3 down")}
	sqsClient := &fakeSQSPublisher{}
	handler := NewHandler(uploader, sqsClient, "test-bucket", "test-queue")

	_, err := handler.PublishForParsing(context.Background(), "uploads/abc.dem", bytes.NewReader(nil))
	if err == nil {
		t.Fatal("expected an error")
	}
	if sqsClient.sent != nil {
		t.Fatal("sqs should not be called when the s3 upload fails")
	}
}

func TestPublishForParsingReturnsErrorOnSQSFailure(t *testing.T) {
	uploader := &fakeUploader{}
	sqsClient := &fakeSQSPublisher{err: errors.New("sqs down")}
	handler := NewHandler(uploader, sqsClient, "test-bucket", "test-queue")

	_, err := handler.PublishForParsing(context.Background(), "uploads/abc.dem", bytes.NewReader(nil))
	if err == nil {
		t.Fatal("expected an error")
	}
}

func multipartUploadRequest(t *testing.T, fieldName, fileName string, content []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write form file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestServeHTTPUploadsFileAndReturnsMessage(t *testing.T) {
	uploader := &fakeUploader{}
	sqsClient := &fakeSQSPublisher{}
	handler := NewHandler(uploader, sqsClient, "test-bucket", "test-queue")

	req := multipartUploadRequest(t, "file", "match.dem", []byte("demo bytes"))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var msg Message
	if err := json.Unmarshal(recorder.Body.Bytes(), &msg); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if msg.Bucket != "test-bucket" {
		t.Fatalf("unexpected bucket: %s", msg.Bucket)
	}
	if string(uploader.uploaded) != "demo bytes" {
		t.Fatalf("unexpected uploaded body: %s", uploader.uploaded)
	}
}

func TestServeHTTPRejectsMissingFileField(t *testing.T) {
	uploader := &fakeUploader{}
	sqsClient := &fakeSQSPublisher{}
	handler := NewHandler(uploader, sqsClient, "test-bucket", "test-queue")

	req := multipartUploadRequest(t, "not-file", "match.dem", []byte("demo bytes"))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestServeHTTPRejectsNonPostMethod(t *testing.T) {
	handler := NewHandler(&fakeUploader{}, &fakeSQSPublisher{}, "test-bucket", "test-queue")

	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}
}

func TestGenerateKeyPreservesExtension(t *testing.T) {
	key, err := GenerateKey("match.dem")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.HasPrefix([]byte(key), []byte("uploads/")) {
		t.Fatalf("expected key to start with uploads/, got %s", key)
	}
	if key[len(key)-4:] != ".dem" {
		t.Fatalf("expected key to preserve .dem extension, got %s", key)
	}
}
