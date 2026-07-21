package veto

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

type Session struct {
	mu        sync.Mutex
	state     State
	sequence  []ActionType
	stepIndex int
	clients   map[*Client]struct{}
}

func newSession(format Format, teamA, teamB string, maps []string) (*Session, error) {
	if teamA == "" || teamB == "" {
		return nil, ErrEmptyTeamName
	}
	if format != FormatBO1 && format != FormatBO3 {
		return nil, ErrInvalidFormat
	}

	seen := make(map[string]struct{}, len(maps))
	for _, m := range maps {
		if _, ok := seen[m]; ok {
			return nil, ErrDuplicateMap
		}
		seen[m] = struct{}{}
	}

	sequence, err := buildSequence(format, len(maps))
	if err != nil {
		return nil, err
	}

	availableMaps := make([]string, len(maps))
	copy(availableMaps, maps)

	session := &Session{
		state: State{
			SessionID:     uuid.New(),
			Format:        format,
			TeamA:         teamA,
			TeamB:         teamB,
			AvailableMaps: availableMaps,
			CurrentTurn:   teamA,
			History:       []VetoEntry{},
		},
		sequence: sequence,
		clients:  make(map[*Client]struct{}),
	}
	session.finalizeIfDone()

	return session, nil
}

func (s *Session) finalizeIfDone() {
	if s.stepIndex < len(s.sequence) {
		return
	}
	if len(s.state.AvailableMaps) == 1 {
		s.state.DeciderMap = s.state.AvailableMaps[0]
	}
	s.state.Completed = true
	s.state.CurrentTurn = ""
}

func (s *Session) ID() uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state.SessionID
}

func (s *Session) Snapshot() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return copyState(s.state)
}

func copyState(state State) State {
	clone := state
	clone.AvailableMaps = append([]string{}, state.AvailableMaps...)
	clone.History = append([]VetoEntry{}, state.History...)
	return clone
}

func (s *Session) AddClient(c *Client) State {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c] = struct{}{}
	return copyState(s.state)
}

func (s *Session) RemoveClient(c *Client) {
	s.mu.Lock()
	delete(s.clients, c)
	s.mu.Unlock()
}

func (s *Session) ApplyAction(team string, action ActionType, mapName string) error {
	s.mu.Lock()

	if s.state.Completed {
		s.mu.Unlock()
		return ErrSessionCompleted
	}
	if team != s.state.TeamA && team != s.state.TeamB {
		s.mu.Unlock()
		return ErrUnknownTeam
	}
	if team != s.state.CurrentTurn {
		s.mu.Unlock()
		return ErrNotYourTurn
	}
	if s.sequence[s.stepIndex] != action {
		s.mu.Unlock()
		return ErrWrongActionType
	}

	mapIndex := -1
	for i, m := range s.state.AvailableMaps {
		if m == mapName {
			mapIndex = i
			break
		}
	}
	if mapIndex == -1 {
		s.mu.Unlock()
		return ErrMapNotAvailable
	}

	s.state.AvailableMaps = append(s.state.AvailableMaps[:mapIndex], s.state.AvailableMaps[mapIndex+1:]...)
	s.state.History = append(s.state.History, VetoEntry{Team: team, Map: mapName, Action: action})
	s.stepIndex++

	if team == s.state.TeamA {
		s.state.CurrentTurn = s.state.TeamB
	} else {
		s.state.CurrentTurn = s.state.TeamA
	}

	s.finalizeIfDone()

	snapshot := copyState(s.state)
	clients := make([]*Client, 0, len(s.clients))
	for c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.Unlock()

	broadcast(clients, snapshot)
	return nil
}

func broadcast(clients []*Client, state State) {
	payload, err := json.Marshal(state)
	if err != nil {
		return
	}
	for _, c := range clients {
		c.Send(payload)
	}
}
