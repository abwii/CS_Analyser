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

const testAllowedOrigin = "http://localhost:5173"
const testTokenTTL = time.Hour

var testSecret = []byte("test-secret-value-that-is-long-enough-for-hs256")

func startTestServer(t *testing.T) (*httptest.Server, *Manager) {
	manager := NewManager()
	mux := http.NewServeMux()
	mux.HandleFunc("/veto/sessions", NewCreateSessionHandler(manager, testSecret, testTokenTTL))
	mux.HandleFunc("/veto/ws", NewWebSocketHandler(manager, testSecret, []string{testAllowedOrigin}))

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, manager
}

func createTestSession(t *testing.T, server *httptest.Server) CreateSessionResponse {
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
	return created
}

func dialTestClientWithHeader(t *testing.T, server *httptest.Server, sessionID string, token string, header http.Header) (*websocket.Conn, *http.Response, error) {
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/veto/ws?sessionId=" + sessionID + "&token=" + token
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if conn != nil {
		t.Cleanup(func() {
			conn.Close()
		})
	}
	return conn, resp, err
}

func dialTestClient(t *testing.T, server *httptest.Server, sessionID string, token string) *websocket.Conn {
	header := http.Header{"Origin": []string{testAllowedOrigin}}
	conn, _, err := dialTestClientWithHeader(t, server, sessionID, token, header)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
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
	created := createTestSession(t, server)

	clientA := dialTestClient(t, server, created.SessionID, created.TeamAToken)
	clientB := dialTestClient(t, server, created.SessionID, created.TeamBToken)

	initialA := readState(t, clientA)
	initialB := readState(t, clientB)

	if initialA.CurrentTurn != "A" || initialB.CurrentTurn != "A" {
		t.Fatalf("expected initial turn to be A for both clients, got %s and %s", initialA.CurrentTurn, initialB.CurrentTurn)
	}
	if initialA.NextAction != ActionBan {
		t.Fatalf("expected next action to be ban, got %s", initialA.NextAction)
	}

	action := ActionRequest{Action: ActionBan, Map: "m1"}
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
	created := createTestSession(t, server)

	clientA := dialTestClient(t, server, created.SessionID, created.TeamAToken)
	clientB := dialTestClient(t, server, created.SessionID, created.TeamBToken)

	readState(t, clientA)
	readState(t, clientB)

	action := ActionRequest{Action: ActionBan, Map: "m1"}
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

func TestWebSocketIgnoresClientSuppliedTeamField(t *testing.T) {
	server, _ := startTestServer(t)
	created := createTestSession(t, server)

	clientA := dialTestClient(t, server, created.SessionID, created.TeamAToken)
	clientB := dialTestClient(t, server, created.SessionID, created.TeamBToken)

	readState(t, clientA)
	readState(t, clientB)

	spoofedPayload := []byte(`{"team":"A","action":"ban","map":"m1"}`)
	if err := clientB.WriteMessage(websocket.TextMessage, spoofedPayload); err != nil {
		t.Fatalf("failed to send spoofed action: %v", err)
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
	if errResp.Error != ErrNotYourTurn.Error() {
		t.Fatalf("expected not your turn error, got %q", errResp.Error)
	}
}

func TestWebSocketRejectsMissingToken(t *testing.T) {
	server, _ := startTestServer(t)
	created := createTestSession(t, server)

	header := http.Header{"Origin": []string{testAllowedOrigin}}
	_, resp, err := dialTestClientWithHeader(t, server, created.SessionID, "", header)
	if err == nil {
		t.Fatalf("expected dial to fail without a token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("expected 401, got %d", status)
	}
}

func TestWebSocketRejectsTokenFromAnotherSession(t *testing.T) {
	server, _ := startTestServer(t)
	created := createTestSession(t, server)
	otherSession := createTestSession(t, server)

	header := http.Header{"Origin": []string{testAllowedOrigin}}
	_, resp, err := dialTestClientWithHeader(t, server, created.SessionID, otherSession.TeamAToken, header)
	if err == nil {
		t.Fatalf("expected dial to fail with a token minted for a different session")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("expected 401, got %d", status)
	}
}

func TestWebSocketRejectsMismatchedOrigin(t *testing.T) {
	server, _ := startTestServer(t)
	created := createTestSession(t, server)

	header := http.Header{"Origin": []string{"https://attacker.example"}}
	_, resp, err := dialTestClientWithHeader(t, server, created.SessionID, created.TeamAToken, header)
	if err == nil {
		t.Fatalf("expected dial to fail with a mismatched origin")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("expected 403, got %d", status)
	}
}
