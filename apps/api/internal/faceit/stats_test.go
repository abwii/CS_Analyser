package faceit

import "testing"

func TestDidPlayerWinTrueWhenInWinningFaction(t *testing.T) {
	history := MatchHistoryItem{
		Results: &MatchResultInfo{Winner: "faction1"},
		Teams: map[string]HistoryFaction{
			"faction1": {Players: []HistoryPlayer{{PlayerID: "p1"}, {PlayerID: "p2"}}},
			"faction2": {Players: []HistoryPlayer{{PlayerID: "p3"}}},
		},
	}

	if !DidPlayerWin(history, "p1") {
		t.Fatalf("expected p1 to be a winner")
	}
	if DidPlayerWin(history, "p3") {
		t.Fatalf("expected p3 not to be a winner")
	}
}

func TestDidPlayerWinFalseWithoutResults(t *testing.T) {
	history := MatchHistoryItem{Teams: map[string]HistoryFaction{}}
	if DidPlayerWin(history, "p1") {
		t.Fatalf("expected false when results are missing")
	}
}

func TestExtractMapReadsFirstRound(t *testing.T) {
	stats := &MatchStats{
		Rounds: []RoundStats{
			{RoundStats: map[string]string{"Map": "de_mirage"}},
		},
	}
	if got := ExtractMap(stats); got != "de_mirage" {
		t.Fatalf("expected de_mirage, got %s", got)
	}
}

func TestExtractMapFallsBackToLowercaseKey(t *testing.T) {
	stats := &MatchStats{
		Rounds: []RoundStats{
			{RoundStats: map[string]string{"map": "de_dust2"}},
		},
	}
	if got := ExtractMap(stats); got != "de_dust2" {
		t.Fatalf("expected de_dust2, got %s", got)
	}
}

func TestExtractMapEmptyWhenMissing(t *testing.T) {
	stats := &MatchStats{Rounds: []RoundStats{{RoundStats: map[string]string{}}}}
	if got := ExtractMap(stats); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
	if got := ExtractMap(nil); got != "" {
		t.Fatalf("expected empty string for nil stats, got %s", got)
	}
}

func TestExtractKDFindsMatchingPlayer(t *testing.T) {
	stats := &MatchStats{
		Rounds: []RoundStats{
			{
				Teams: []TeamStats{
					{
						Players: []PlayerRoundStats{
							{PlayerID: "other", PlayerStats: map[string]string{"Kills": "5", "Deaths": "10"}},
							{PlayerID: "target", PlayerStats: map[string]string{"Kills": "23", "Deaths": "12"}},
						},
					},
				},
			},
		},
	}

	kills, deaths, ok := ExtractKD(stats, "target")
	if !ok {
		t.Fatalf("expected ok to be true")
	}
	if kills != 23 || deaths != 12 {
		t.Fatalf("expected 23/12, got %d/%d", kills, deaths)
	}
}

func TestExtractKDFalseWhenPlayerMissing(t *testing.T) {
	stats := &MatchStats{Rounds: []RoundStats{{Teams: []TeamStats{}}}}
	_, _, ok := ExtractKD(stats, "target")
	if ok {
		t.Fatalf("expected ok to be false")
	}
}

func TestAggregateByMapGroupsAndComputesRates(t *testing.T) {
	results := []MatchResult{
		{Map: "de_mirage", Won: true, Kills: 20, Deaths: 10, HasKD: true},
		{Map: "de_mirage", Won: false, Kills: 10, Deaths: 20, HasKD: true},
		{Map: "de_dust2", Won: true, Kills: 25, Deaths: 15, HasKD: true},
		{Map: "", Won: true, Kills: 99, Deaths: 1, HasKD: true},
	}

	stats := AggregateByMap(results)

	if len(stats) != 2 {
		t.Fatalf("expected 2 maps, got %d", len(stats))
	}

	if stats[0].Map != "de_dust2" || stats[1].Map != "de_mirage" {
		t.Fatalf("expected alphabetical order, got %+v", stats)
	}

	mirage := stats[1]
	if mirage.MatchesPlayed != 2 || mirage.Wins != 1 {
		t.Fatalf("unexpected mirage aggregate: %+v", mirage)
	}
	if mirage.WinRate != 50.0 {
		t.Fatalf("expected 50 percent win rate, got %v", mirage.WinRate)
	}
	if mirage.AvgKills != 15.0 || mirage.AvgDeaths != 15.0 {
		t.Fatalf("expected avg kills/deaths of 15, got %v/%v", mirage.AvgKills, mirage.AvgDeaths)
	}
}

func TestAggregateByMapSkipsSamplesWithoutKD(t *testing.T) {
	results := []MatchResult{
		{Map: "de_nuke", Won: true, HasKD: false},
	}

	stats := AggregateByMap(results)
	if len(stats) != 1 {
		t.Fatalf("expected 1 map, got %d", len(stats))
	}
	if stats[0].AvgKills != 0 || stats[0].AvgDeaths != 0 {
		t.Fatalf("expected zero averages without kd samples, got %+v", stats[0])
	}
}

func TestAggregateByMapEmptyInput(t *testing.T) {
	stats := AggregateByMap(nil)
	if len(stats) != 0 {
		t.Fatalf("expected no stats, got %+v", stats)
	}
}
