package veto

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/wbach/cs-analyser/apps/api/internal/auth"
)

type CreateSessionRequest struct {
	Format Format   `json:"format"`
	TeamA  string   `json:"teamA"`
	TeamB  string   `json:"teamB"`
	Maps   []string `json:"maps"`
}

type CreateSessionResponse struct {
	SessionID  string `json:"sessionId"`
	TeamAToken string `json:"teamAToken"`
	TeamBToken string `json:"teamBToken"`
}

type ActionRequest struct {
	Action ActionType `json:"action"`
	Map    string     `json:"map"`
}

type errorMessage struct {
	Error string `json:"error"`
}

func NewCreateSessionHandler(manager *Manager, secret []byte, tokenTTL time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, err := manager.CreateSession(req.Format, req.TeamA, req.TeamB, req.Maps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sessionID := session.ID().String()

		teamAToken, err := auth.NewVetoToken(secret, sessionID, req.TeamA, tokenTTL)
		if err != nil {
			log.Printf("failed to sign team token: %v", err)
			http.Error(w, "failed to issue session tokens", http.StatusInternalServerError)
			return
		}
		teamBToken, err := auth.NewVetoToken(secret, sessionID, req.TeamB, tokenTTL)
		if err != nil {
			log.Printf("failed to sign team token: %v", err)
			http.Error(w, "failed to issue session tokens", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := CreateSessionResponse{SessionID: sessionID, TeamAToken: teamAToken, TeamBToken: teamBToken}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("failed to write create session response: %v", err)
		}
	}
}

func NewWebSocketHandler(manager *Manager, secret []byte, allowedOrigin string) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == allowedOrigin
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		sessionIDParam := r.URL.Query().Get("sessionId")
		sessionID, err := uuid.Parse(sessionIDParam)
		if err != nil {
			http.Error(w, "invalid session id", http.StatusBadRequest)
			return
		}

		session, ok := manager.Get(sessionID)
		if !ok {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		claims, err := auth.Parse(secret, r.URL.Query().Get("token"))
		if err != nil || claims.Scope != auth.ScopeVeto || claims.SessionID != sessionIDParam {
			http.Error(w, "invalid or missing session token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade failed: %v", err)
			return
		}

		client := newClient(conn, claims.Team)
		go client.writePump()

		snapshot := session.AddClient(client)
		sendState(client, snapshot)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var actionReq ActionRequest
			if err := json.Unmarshal(message, &actionReq); err != nil {
				sendError(client, "invalid message format")
				continue
			}

			if err := session.ApplyAction(client.Team, actionReq.Action, actionReq.Map); err != nil {
				sendError(client, err.Error())
			}
		}

		session.RemoveClient(client)
		client.Close()
	}
}

func sendState(c *Client, state State) {
	payload, err := json.Marshal(state)
	if err != nil {
		log.Printf("failed to marshal state: %v", err)
		return
	}
	c.Send(payload)
}

func sendError(c *Client, message string) {
	payload, err := json.Marshal(errorMessage{Error: message})
	if err != nil {
		log.Printf("failed to marshal error message: %v", err)
		return
	}
	c.Send(payload)
}
