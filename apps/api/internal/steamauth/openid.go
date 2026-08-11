package steamauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const steamOpenIDEndpoint = "https://steamcommunity.com/openid/login"
const openIDNamespace = "http://specs.openid.net/auth/2.0"
const identifierSelect = "http://specs.openid.net/auth/2.0/identifier_select"

var ErrNotAuthenticated = errors.New("steam did not confirm the openid response")

var claimedIDPattern = regexp.MustCompile(`^https://steamcommunity\.com/openid/id/(\d+)$`)

func BuildLoginURL(returnTo, realm string) string {
	values := url.Values{
		"openid.ns":         {openIDNamespace},
		"openid.mode":       {"checkid_setup"},
		"openid.return_to":  {returnTo},
		"openid.realm":      {realm},
		"openid.identity":   {identifierSelect},
		"openid.claimed_id": {identifierSelect},
	}
	return steamOpenIDEndpoint + "?" + values.Encode()
}

type Verifier struct {
	endpoint   string
	httpClient *http.Client
}

func NewVerifier(httpClient *http.Client) *Verifier {
	return &Verifier{endpoint: steamOpenIDEndpoint, httpClient: httpClient}
}

func (v *Verifier) Verify(ctx context.Context, callbackQuery url.Values) (string, error) {
	steamID, err := extractSteamID(callbackQuery.Get("openid.claimed_id"))
	if err != nil {
		return "", err
	}

	verifyParams := url.Values{}
	for key, values := range callbackQuery {
		verifyParams[key] = values
	}
	verifyParams.Set("openid.mode", "check_authentication")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, strings.NewReader(verifyParams.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("steam openid verification failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if !isValidResponse(string(body)) {
		return "", ErrNotAuthenticated
	}

	return steamID, nil
}

func isValidResponse(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == "is_valid:true" {
			return true
		}
	}
	return false
}

func extractSteamID(claimedID string) (string, error) {
	matches := claimedIDPattern.FindStringSubmatch(claimedID)
	if matches == nil {
		return "", fmt.Errorf("unexpected openid.claimed_id format: %q", claimedID)
	}
	return matches[1], nil
}
