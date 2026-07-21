package veto

import (
	"errors"

	"github.com/google/uuid"
)

type Format string

const (
	FormatBO1 Format = "BO1"
	FormatBO3 Format = "BO3"
)

type ActionType string

const (
	ActionBan  ActionType = "ban"
	ActionPick ActionType = "pick"
)

var (
	ErrInvalidFormat    = errors.New("invalid format")
	ErrNotEnoughMaps    = errors.New("not enough maps for this format")
	ErrDuplicateMap     = errors.New("duplicate map in pool")
	ErrEmptyTeamName    = errors.New("team names must not be empty")
	ErrSessionCompleted = errors.New("session already completed")
	ErrUnknownTeam      = errors.New("unknown team")
	ErrNotYourTurn      = errors.New("not this team's turn")
	ErrWrongActionType  = errors.New("action type does not match the expected step")
	ErrMapNotAvailable  = errors.New("map not available")
)

type VetoEntry struct {
	Team   string     `json:"team"`
	Map    string     `json:"map"`
	Action ActionType `json:"action"`
}

type State struct {
	SessionID     uuid.UUID   `json:"sessionId"`
	Format        Format      `json:"format"`
	TeamA         string      `json:"teamA"`
	TeamB         string      `json:"teamB"`
	AvailableMaps []string    `json:"availableMaps"`
	CurrentTurn   string      `json:"currentTurn"`
	NextAction    ActionType  `json:"nextAction,omitempty"`
	History       []VetoEntry `json:"history"`
	DeciderMap    string      `json:"deciderMap,omitempty"`
	Completed     bool        `json:"completed"`
}

func buildSequence(format Format, mapCount int) ([]ActionType, error) {
	if mapCount < 1 {
		return nil, ErrNotEnoughMaps
	}

	totalActions := mapCount - 1

	switch format {
	case FormatBO1:
		sequence := make([]ActionType, totalActions)
		for i := range sequence {
			sequence[i] = ActionBan
		}
		return sequence, nil
	case FormatBO3:
		picks := 2
		if totalActions < picks {
			return nil, ErrNotEnoughMaps
		}
		bansNeeded := totalActions - picks
		before := bansNeeded / 2
		after := bansNeeded - before

		sequence := make([]ActionType, 0, totalActions)
		for i := 0; i < before; i++ {
			sequence = append(sequence, ActionBan)
		}
		for i := 0; i < picks; i++ {
			sequence = append(sequence, ActionPick)
		}
		for i := 0; i < after; i++ {
			sequence = append(sequence, ActionBan)
		}
		return sequence, nil
	default:
		return nil, ErrInvalidFormat
	}
}
