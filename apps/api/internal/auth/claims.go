package auth

import "github.com/golang-jwt/jwt/v5"

const ScopeAdmin = "admin"
const ScopeVeto = "veto"
const ScopePlayer = "player"

type Claims struct {
	jwt.RegisteredClaims
	Scope     string `json:"scope"`
	Team      string `json:"team,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	SteamID   string `json:"steamId,omitempty"`
}
