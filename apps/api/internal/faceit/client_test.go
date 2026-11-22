package faceit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{baseURL: server.URL, apiKey: "test-key", httpClient: server.Client()}
}

func TestFindPlayerBySteamIDSendsAuthAndQuery(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		if r.URL.Path != "/players" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("game_player_id"); got != "76561198000000001" {
			t.Fatalf("unexpected game_player_id: %s", got)
		}
		if got := r.URL.Query().Get("game"); got != "cs2" {
			t.Fatalf("unexpected game: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"player_id":"abc123","nickname":"tester"}`))
	})

	player, err := client.FindPlayerBySteamID(context.Background(), "76561198000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.PlayerID != "abc123" || player.Nickname != "tester" {
		t.Fatalf("unexpected player: %+v", player)
	}
}

func TestFindPlayerBySteamIDNotFound(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.FindPlayerBySteamID(context.Background(), "76561198000000001")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMatchHistoryParsesItems(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/players/abc123/history" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Fatalf("unexpected limit: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"match_id":"m1","results":{"winner":"faction1"},"teams":{"faction1":{"players":[{"player_id":"abc123"}]}}}]}`))
	})

	items, err := client.MatchHistory(context.Background(), "abc123", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].MatchID != "m1" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if !DidPlayerWin(items[0], "abc123") {
		t.Fatalf("expected abc123 to be the winner")
	}
}

func TestMatchStatsParsesRoundStats(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/matches/m1/stats" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rounds":[{"round_stats":{"Map":"de_mirage"},"teams":[{"team_id":"faction1","players":[{"player_id":"abc123","player_stats":{"Kills":"20","Deaths":"10"}}]}]}]}`))
	})

	stats, err := client.MatchStats(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ExtractMap(stats) != "de_mirage" {
		t.Fatalf("expected de_mirage, got %s", ExtractMap(stats))
	}
	kills, deaths, ok := ExtractKD(stats, "abc123")
	if !ok || kills != 20 || deaths != 10 {
		t.Fatalf("unexpected kd: %d/%d ok=%v", kills, deaths, ok)
	}
}

func TestGetMatchParsesDemoURL(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/matches/m1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"match_id":"m1","demo_url":["https://demos.faceit-cdn.net/m1.dem"]}`))
	})

	details, err := client.GetMatch(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(details.DemoURL) != 1 || details.DemoURL[0] != "https://demos.faceit-cdn.net/m1.dem" {
		t.Fatalf("unexpected demo url: %+v", details.DemoURL)
	}
}

func TestGetMatchHandlesMissingDemoURL(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"match_id":"m1","demo_url":[]}`))
	})

	details, err := client.GetMatch(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(details.DemoURL) != 0 {
		t.Fatalf("expected no demo url, got %+v", details.DemoURL)
	}
}

func TestGetReturnsErrorOnServerError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	})

	_, err := client.MatchStats(context.Background(), "m1")
	if err == nil {
		t.Fatalf("expected an error")
	}
}
