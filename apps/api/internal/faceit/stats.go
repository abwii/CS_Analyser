package faceit

import "sort"

type MatchResult struct {
	Map    string
	Won    bool
	Kills  int
	Deaths int
	HasKD  bool
}

type MapStats struct {
	Map           string  `json:"map"`
	MatchesPlayed int     `json:"matchesPlayed"`
	Wins          int     `json:"wins"`
	WinRate       float64 `json:"winRate"`
	AvgKills      float64 `json:"avgKills"`
	AvgDeaths     float64 `json:"avgDeaths"`
}

func DidPlayerWin(history MatchHistoryItem, faceitPlayerID string) bool {
	if history.Results == nil {
		return false
	}
	faction, ok := history.Teams[history.Results.Winner]
	if !ok {
		return false
	}
	for _, player := range faction.Players {
		if player.PlayerID == faceitPlayerID {
			return true
		}
	}
	return false
}

func ExtractMap(stats *MatchStats) string {
	if stats == nil || len(stats.Rounds) == 0 {
		return ""
	}
	round := stats.Rounds[0]
	if value, ok := round.RoundStats["Map"]; ok && value != "" {
		return value
	}
	if value, ok := round.RoundStats["map"]; ok && value != "" {
		return value
	}
	return ""
}

func ExtractKD(stats *MatchStats, faceitPlayerID string) (kills int, deaths int, ok bool) {
	if stats == nil || len(stats.Rounds) == 0 {
		return 0, 0, false
	}
	for _, team := range stats.Rounds[0].Teams {
		for _, player := range team.Players {
			if player.PlayerID != faceitPlayerID {
				continue
			}
			kills, killsOk := parseIntStat(player.PlayerStats, "Kills")
			deaths, deathsOk := parseIntStat(player.PlayerStats, "Deaths")
			if killsOk && deathsOk {
				return kills, deaths, true
			}
			return 0, 0, false
		}
	}
	return 0, 0, false
}

func parseIntStat(stats map[string]string, key string) (int, bool) {
	raw, ok := stats[key]
	if !ok {
		return 0, false
	}
	value := 0
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, false
		}
		value = value*10 + int(c-'0')
	}
	return value, true
}

func AggregateByMap(results []MatchResult) []MapStats {
	type accumulator struct {
		matches   int
		wins      int
		kills     int
		deaths    int
		kdSamples int
	}

	byMap := make(map[string]*accumulator)
	order := make([]string, 0)

	for _, result := range results {
		if result.Map == "" {
			continue
		}
		acc, exists := byMap[result.Map]
		if !exists {
			acc = &accumulator{}
			byMap[result.Map] = acc
			order = append(order, result.Map)
		}
		acc.matches++
		if result.Won {
			acc.wins++
		}
		if result.HasKD {
			acc.kills += result.Kills
			acc.deaths += result.Deaths
			acc.kdSamples++
		}
	}

	stats := make([]MapStats, 0, len(order))
	for _, mapName := range order {
		acc := byMap[mapName]
		mapStats := MapStats{
			Map:           mapName,
			MatchesPlayed: acc.matches,
			Wins:          acc.wins,
			WinRate:       float64(acc.wins) / float64(acc.matches) * 100,
		}
		if acc.kdSamples > 0 {
			mapStats.AvgKills = float64(acc.kills) / float64(acc.kdSamples)
			mapStats.AvgDeaths = float64(acc.deaths) / float64(acc.kdSamples)
		}
		stats = append(stats, mapStats)
	}

	sort.Slice(stats, func(i, j int) bool { return stats[i].Map < stats[j].Map })
	return stats
}
