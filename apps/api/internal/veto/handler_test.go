package veto

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func startTestServer(t *testing.T) (*httptest.Server, *Manager) {
	manager := NewManager()
	mux := http.NewServeMux()
	mux.HandleFunc("/veto/sessions", NewCreateSessionHandler(manager))
	mux.HandleFunc("/veto/ws", NewWebSocketHandler(manager))

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, manager
}

func createTestSession(t *testing.T, server *httptest.Server) string {
	body := CreateSessionRequest{
		Format: FormatBO1,
		TeamA:  "A",
		TeamB:  "B",
		Maps:   []string{"m1", "m2", "m3"},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(server.URL+"/veto/sessions", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var created CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return created.SessionID
}

func dialTestClient(t *testing.T, server *httptest.Server, sessionID string) *websocket.Conn {
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/veto/ws?sessionId=" + sessionID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	t.Cleanup(func() {
		conn.Close()
	})
	return conn
}

func readState(t *testing.T, conn *websocket.Conn) State {
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	var state State
	if err := json.Unmarshal(message, &state); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}
	return state
}

func TestWebSocketBroadcastsStateToAllClients(t *testing.T) {
	server, _ := startTestServer(t)
	sessionID := createTestSession(t, server)

	clientA := dialTestClient(t, server, sessionID)
	clientB := dialTestClient(t, server, sessionID)

	initialA := readState(t, clientA)
	initialB := readState(t, clientB)

	if initialA.CurrentTurn != "A" || initialB.CurrentTurn != "A" {
		t.Fatalf("expected initial turn to be A for both clients, got %s and %s", initialA.CurrentTurn, initialB.CurrentTurn)
	}
	if initialA.NextAction != ActionBan {
		t.Fatalf("expected next action to be ban, got %s", initialA.NextAction)
	}

	action := ActionRequest{Team: "A", Action: ActionBan, Map: "m1"}
	payload, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal action: %v", err)
	}
	if err := clientA.WriteMessage(websocket.TextMessage, payload); err != nil {
		t.Fatalf("failed to send action: %v", err)
	}

	stateFromA := readState(t, clientA)
	stateFromB := readState(t, clientB)

	for _, state := range []State{stateFromA, stateFromB} {
		if state.CurrentTurn != "B" {
			t.Fatalf("expected turn to switch to B, got %s", state.CurrentTurn)
		}
		if len(state.History) != 1 || state.History[0].Map != "m1" {
			t.Fatalf("expected history with m1 banned, got %+v", state.History)
		}
		for _, m := range state.AvailableMaps {
			if m == "m1" {
				t.Fatalf("expected m1 to be removed from available maps")
			}
		}
	}
}

func TestWebSocketRejectsInvalidActionWithoutBroadcast(t *testing.T) {
	server, _ := startTestServer(t)
	sessionID := createTestSession(t, server)

	clientA := dialTestClient(t, server, sessionID)
	clientB := dialTestClient(t, server, sessionID)

	readState(t, clientA)
	readState(t, clientB)

	action := ActionRequest{Team: "B", Action: ActionBan, Map: "m1"}
	payload, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal action: %v", err)
	}
	if err := clientB.WriteMessage(websocket.TextMessage, payload); err != nil {
		t.Fatalf("failed to send action: %v", err)
	}

	if err := clientB.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	_, message, err := clientB.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read error message: %v", err)
	}
	var errResp errorMessage
	if err := json.Unmarshal(message, &errResp); err != nil {
		t.Fatalf("failed to unmarshal error message: %v", err)
	}
	if errResp.Error == "" {
		t.Fatalf("expected a non-empty error message")
	}

	if err := clientA.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	if _, _, err := clientA.ReadMessage(); err == nil {
		t.Fatalf("expected no broadcast to client A after an invalid action")
	}
}
