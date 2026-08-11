package steamauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const summaryAPIBaseURL = "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/"

var ErrPlayerNotFound = errors.New("steam player not found")

type PlayerSummary struct {
	SteamID     string `json:"steamId"`
	PersonaName string `json:"personaName"`
	AvatarURL   string `json:"avatarUrl"`
}

type SummaryClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewSummaryClient(apiKey string) *SummaryClient {
	return &SummaryClient{
		baseURL:    summaryAPIBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *SummaryClient) Configured() bool {
	return c.apiKey != ""
}

type playerSummaryResponse struct {
	Response struct {
		Players []struct {
			SteamID      string `json:"steamid"`
			PersonaName  string `json:"personaname"`
			AvatarMedium string `json:"avatarmedium"`
		} `json:"players"`
	} `json:"response"`
}

func (c *SummaryClient) GetPlayerSummary(ctx context.Context, steamID string) (*PlayerSummary, error) {
	query := url.Values{}
	query.Set("key", c.apiKey)
	query.Set("steamids", steamID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam web api error: status %d: %s", resp.StatusCode, string(body))
	}

	var parsed playerSummaryResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Response.Players) == 0 {
		return nil, ErrPlayerNotFound
	}

	player := parsed.Response.Players[0]
	return &PlayerSummary{
		SteamID:     player.SteamID,
		PersonaName: player.PersonaName,
		AvatarURL:   player.AvatarMedium,
	}, nil
}
