package veto

import (
	"sync"
	"testing"
)

func TestBuildSequenceBO1(t *testing.T) {
	sequence, err := buildSequence(FormatBO1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sequence) != 6 {
		t.Fatalf("expected 6 steps, got %d", len(sequence))
	}
	for _, action := range sequence {
		if action != ActionBan {
			t.Fatalf("expected all bans, got %s", action)
		}
	}
}

func TestBuildSequenceBO3(t *testing.T) {
	sequence, err := buildSequence(FormatBO3, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []ActionType{ActionBan, ActionBan, ActionPick, ActionPick, ActionBan, ActionBan}
	if len(sequence) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(sequence))
	}
	for i, action := range sequence {
		if action != expected[i] {
			t.Fatalf("step %d: expected %s, got %s", i, expected[i], action)
		}
	}
}

func TestBuildSequenceBO3TooFewMaps(t *testing.T) {
	if _, err := buildSequence(FormatBO3, 2); err != ErrNotEnoughMaps {
		t.Fatalf("expected ErrNotEnoughMaps, got %v", err)
	}
}

func TestNewSessionRejectsInvalidInput(t *testing.T) {
	if _, err := newSession(FormatBO1, "", "B", []string{"a", "b"}); err != ErrEmptyTeamName {
		t.Fatalf("expected ErrEmptyTeamName, got %v", err)
	}
	if _, err := newSession("BO5", "A", "B", []string{"a", "b"}); err != ErrInvalidFormat {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
	if _, err := newSession(FormatBO1, "A", "B", []string{"a", "a"}); err != ErrDuplicateMap {
		t.Fatalf("expected ErrDuplicateMap, got %v", err)
	}
}

func TestApplyActionRejectsWrongTurn(t *testing.T) {
	session, err := newSession(FormatBO1, "A", "B", []string{"m1", "m2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := session.ApplyAction("B", ActionBan, "m1"); err != ErrNotYourTurn {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}
}

func TestApplyActionRejectsUnknownTeam(t *testing.T) {
	session, err := newSession(FormatBO1, "A", "B", []string{"m1", "m2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := session.ApplyAction("C", ActionBan, "m1"); err != ErrUnknownTeam {
		t.Fatalf("expected ErrUnknownTeam, got %v", err)
	}
}

func TestApplyActionRejectsWrongActionType(t *testing.T) {
	session, err := newSession(FormatBO3, "A", "B", []string{"m1", "m2", "m3", "m4", "m5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := session.ApplyAction("A", ActionPick, "m1"); err != ErrWrongActionType {
		t.Fatalf("expected ErrWrongActionType, got %v", err)
	}
}

func TestApplyActionRejectsUnavailableMap(t *testing.T) {
	session, err := newSession(FormatBO1, "A", "B", []string{"m1", "m2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := session.ApplyAction("A", ActionBan, "m3"); err != ErrMapNotAvailable {
		t.Fatalf("expected ErrMapNotAvailable, got %v", err)
	}
}

func TestApplyActionRejectsAfterCompletion(t *testing.T) {
	session, err := newSession(FormatBO1, "A", "B", []string{"m1", "m2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := session.ApplyAction("A", ActionBan, "m1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state := session.Snapshot()
	if !state.Completed {
		t.Fatalf("expected session to be completed")
	}
	if state.DeciderMap != "m2" {
		t.Fatalf("expected decider map m2, got %s", state.DeciderMap)
	}

	if err := session.ApplyAction("B", ActionBan, "m2"); err != ErrSessionCompleted {
		t.Fatalf("expected ErrSessionCompleted, got %v", err)
	}
}

func TestApplyActionFullBO3Sequence(t *testing.T) {
	maps := []string{"m1", "m2", "m3", "m4", "m5"}
	session, err := newSession(FormatBO3, "A", "B", maps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	steps := []struct {
		team    string
		action  ActionType
		mapName string
	}{
		{"A", ActionBan, "m1"},
		{"B", ActionPick, "m2"},
		{"A", ActionPick, "m3"},
		{"B", ActionBan, "m4"},
	}

	for _, step := range steps {
		if err := session.ApplyAction(step.team, step.action, step.mapName); err != nil {
			t.Fatalf("unexpected error on step %+v: %v", step, err)
		}
	}

	state := session.Snapshot()
	if !state.Completed {
		t.Fatalf("expected session to be completed")
	}
	if state.DeciderMap != "m5" {
		t.Fatalf("expected decider map m5, got %s", state.DeciderMap)
	}
	if len(state.History) != 4 {
		t.Fatalf("expected 4 history entries, got %d", len(state.History))
	}
	if state.CurrentTurn != "" {
		t.Fatalf("expected empty turn after completion, got %s", state.CurrentTurn)
	}
}

func TestApplyActionConcurrentOnlyOneWinnerPerStep(t *testing.T) {
	session, err := newSession(FormatBO1, "A", "B", []string{"m1", "m2", "m3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	var successCount int
	var mu sync.Mutex

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := session.ApplyAction("A", ActionBan, "m1"); err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful action, got %d", successCount)
	}
}
