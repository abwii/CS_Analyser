package faceit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultBaseURL = "https://open.faceit.com/data/v4"
const defaultGameID = "cs2"

var ErrNotFound = errors.New("faceit resource not found")

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Configured() bool {
	return c.apiKey != ""
}

type Player struct {
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
}

type MatchHistoryItem struct {
	MatchID string                    `json:"match_id"`
	Results *MatchResultInfo          `json:"results"`
	Teams   map[string]HistoryFaction `json:"teams"`
}

type MatchResultInfo struct {
	Winner string `json:"winner"`
}

type HistoryFaction struct {
	Players []HistoryPlayer `json:"players"`
}

type HistoryPlayer struct {
	PlayerID string `json:"player_id"`
}

type MatchDetails struct {
	MatchID string   `json:"match_id"`
	DemoURL []string `json:"demo_url"`
}

type matchHistoryResponse struct {
	Items []MatchHistoryItem `json:"items"`
}

type MatchStats struct {
	Rounds []RoundStats `json:"rounds"`
}

type RoundStats struct {
	RoundStats map[string]string `json:"round_stats"`
	Teams      []TeamStats       `json:"teams"`
}

type TeamStats struct {
	TeamID  string             `json:"team_id"`
	Players []PlayerRoundStats `json:"players"`
}

type PlayerRoundStats struct {
	PlayerID    string            `json:"player_id"`
	PlayerStats map[string]string `json:"player_stats"`
}

func (c *Client) FindPlayerBySteamID(ctx context.Context, steamID string) (*Player, error) {
	query := url.Values{}
	query.Set("game", defaultGameID)
	query.Set("game_player_id", steamID)

	var player Player
	if err := c.get(ctx, "/players", query, &player); err != nil {
		return nil, err
	}
	if player.PlayerID == "" {
		return nil, ErrNotFound
	}
	return &player, nil
}

func (c *Client) MatchHistory(ctx context.Context, playerID string, limit int) ([]MatchHistoryItem, error) {
	query := url.Values{}
	query.Set("game", defaultGameID)
	query.Set("offset", "0")
	query.Set("limit", strconv.Itoa(limit))

	var resp matchHistoryResponse
	if err := c.get(ctx, fmt.Sprintf("/players/%s/history", playerID), query, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) MatchStats(ctx context.Context, matchID string) (*MatchStats, error) {
	var stats MatchStats
	if err := c.get(ctx, fmt.Sprintf("/matches/%s/stats", matchID), nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetMatch fetches match details, notably DemoURL. In practice FACEIT leaves
// this empty for most matchmaking matches despite documenting the field, so
// callers must treat an empty slice as "no demo available", not an error.
func (c *Client) GetMatch(ctx context.Context, matchID string) (*MatchDetails, error) {
	var details MatchDetails
	if err := c.get(ctx, fmt.Sprintf("/matches/%s", matchID), nil, &details); err != nil {
		return nil, err
	}
	return &details, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, dest interface{}) error {
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("faceit api error: status %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, dest)
}
