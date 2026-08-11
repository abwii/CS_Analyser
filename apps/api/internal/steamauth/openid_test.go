package steamauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBuildLoginURLIncludesRequiredParams(t *testing.T) {
	loginURL := BuildLoginURL("http://localhost:8080/auth/steam/callback", "http://localhost:8080")

	parsed, err := url.Parse(loginURL)
	if err != nil {
		t.Fatalf("unexpected error parsing login url: %v", err)
	}
	if !strings.HasPrefix(loginURL, steamOpenIDEndpoint+"?") {
		t.Fatalf("expected login url to target steam openid endpoint, got %s", loginURL)
	}

	query := parsed.Query()
	if query.Get("openid.mode") != "checkid_setup" {
		t.Fatalf("unexpected openid.mode: %s", query.Get("openid.mode"))
	}
	if query.Get("openid.return_to") != "http://localhost:8080/auth/steam/callback" {
		t.Fatalf("unexpected openid.return_to: %s", query.Get("openid.return_to"))
	}
	if query.Get("openid.realm") != "http://localhost:8080" {
		t.Fatalf("unexpected openid.realm: %s", query.Get("openid.realm"))
	}
	if query.Get("openid.identity") != identifierSelect {
		t.Fatalf("unexpected openid.identity: %s", query.Get("openid.identity"))
	}
}

func testVerifier(t *testing.T, handler http.HandlerFunc) *Verifier {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Verifier{endpoint: server.URL, httpClient: server.Client()}
}

func callbackQuery(claimedID string) url.Values {
	return url.Values{
		"openid.ns":             {openIDNamespace},
		"openid.mode":           {"id_res"},
		"openid.claimed_id":     {claimedID},
		"openid.identity":       {claimedID},
		"openid.return_to":      {"http://localhost:8080/auth/steam/callback"},
		"openid.response_nonce": {"2026-07-28T00:00:00Zsomenonce"},
		"openid.assoc_handle":   {"handle"},
		"openid.signed":         {"signed,fields"},
		"openid.sig":            {"signature"},
	}
}

func TestVerifySendsCheckAuthenticationAndReturnsSteamID(t *testing.T) {
	verifier := testVerifier(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse verification request body: %v", err)
		}
		if got := r.PostForm.Get("openid.mode"); got != "check_authentication" {
			t.Fatalf("expected openid.mode=check_authentication in verification request, got %s", got)
		}
		if got := r.PostForm.Get("openid.sig"); got != "signature" {
			t.Fatalf("expected original openid.sig to be forwarded, got %s", got)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ns:http://specs.openid.net/auth/2.0\nis_valid:true\n"))
	})

	steamID, err := verifier.Verify(context.Background(), callbackQuery("https://steamcommunity.com/openid/id/76561198158653120"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if steamID != "76561198158653120" {
		t.Fatalf("unexpected steamID: %s", steamID)
	}
}

func TestVerifyRejectsWhenSteamSaysInvalid(t *testing.T) {
	verifier := testVerifier(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ns:http://specs.openid.net/auth/2.0\nis_valid:false\n"))
	})

	_, err := verifier.Verify(context.Background(), callbackQuery("https://steamcommunity.com/openid/id/76561198158653120"))
	if err == nil {
		t.Fatal("expected an error when steam reports is_valid:false")
	}
}

func TestVerifyRejectsMalformedClaimedID(t *testing.T) {
	verifier := testVerifier(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("steam should not be contacted when claimed_id is malformed")
	})

	_, err := verifier.Verify(context.Background(), callbackQuery("https://evil.example.com/openid/id/76561198158653120"))
	if err == nil {
		t.Fatal("expected an error for a claimed_id outside steamcommunity.com")
	}
}

func TestExtractSteamIDParsesClaimedID(t *testing.T) {
	steamID, err := extractSteamID("https://steamcommunity.com/openid/id/76561198158653120")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if steamID != "76561198158653120" {
		t.Fatalf("unexpected steamID: %s", steamID)
	}
}

func TestExtractSteamIDRejectsUnexpectedFormat(t *testing.T) {
	if _, err := extractSteamID("not-a-url"); err == nil {
		t.Fatal("expected an error for malformed claimed_id")
	}
}
