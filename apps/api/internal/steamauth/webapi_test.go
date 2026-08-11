package steamauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testSummaryClient(t *testing.T, handler http.HandlerFunc) *SummaryClient {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &SummaryClient{baseURL: server.URL, apiKey: "test-key", httpClient: server.Client()}
}

func TestGetPlayerSummarySendsKeyAndSteamID(t *testing.T) {
	client := testSummaryClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("key"); got != "test-key" {
			t.Fatalf("unexpected key: %s", got)
		}
		if got := r.URL.Query().Get("steamids"); got != "76561198158653120" {
			t.Fatalf("unexpected steamids: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"players":[{"steamid":"76561198158653120","personaname":"abwi","avatarmedium":"https://avatars.steamstatic.com/abc_medium.jpg"}]}}`))
	})

	summary, err := client.GetPlayerSummary(context.Background(), "76561198158653120")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.PersonaName != "abwi" {
		t.Fatalf("unexpected personaName: %s", summary.PersonaName)
	}
	if summary.AvatarURL != "https://avatars.steamstatic.com/abc_medium.jpg" {
		t.Fatalf("unexpected avatarUrl: %s", summary.AvatarURL)
	}
}

func TestGetPlayerSummaryNotFoundWhenPlayersEmpty(t *testing.T) {
	client := testSummaryClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	})

	_, err := client.GetPlayerSummary(context.Background(), "76561198158653120")
	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestGetPlayerSummaryErrorsOnNonOKStatus(t *testing.T) {
	client := testSummaryClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden"))
	})

	if _, err := client.GetPlayerSummary(context.Background(), "76561198158653120"); err == nil {
		t.Fatal("expected an error on non-200 status")
	}
}

func TestConfiguredReflectsAPIKeyPresence(t *testing.T) {
	if NewSummaryClient("").Configured() {
		t.Fatal("expected Configured() to be false with an empty api key")
	}
	if !NewSummaryClient("some-key").Configured() {
		t.Fatal("expected Configured() to be true with a non-empty api key")
	}
}
