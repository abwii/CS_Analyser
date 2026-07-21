pub type PlayerId = u64;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct KillEvent {
    pub round: u32,
    pub attacker_id: Option<PlayerId>,
    pub attacker_team: Option<u8>,
    pub victim_id: PlayerId,
    pub victim_team: u8,
    pub assister_id: Option<PlayerId>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct DamageEvent {
    pub round: u32,
    pub attacker_id: PlayerId,
    pub attacker_team: u8,
    pub victim_id: PlayerId,
    pub victim_team: u8,
    pub damage: u32,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct RoundEndEvent {
    pub round: u32,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MatchEvent {
    Kill(KillEvent),
    Damage(DamageEvent),
    RoundEnd(RoundEndEvent),
}

#[derive(Debug, Clone, PartialEq)]
pub struct PlayerStats {
    pub player_id: PlayerId,
    pub kpr: f64,
    pub dpr: f64,
    pub kast: f64,
    pub impact: f64,
    pub adr: f64,
}

#[derive(Debug, Clone, PartialEq)]
pub struct NormalizedPlayerStats {
    pub player_id: PlayerId,
    pub kpr: f64,
    pub dpr: f64,
    pub kast: f64,
    pub impact: f64,
    pub adr: f64,
}

pub fn compute_player_stats(events: &[MatchEvent]) -> Vec<PlayerStats> {
    let kills: Vec<KillEvent> = events
        .iter()
        .filter_map(|e| match e {
            MatchEvent::Kill(k) => Some(*k),
            _ => None,
        })
        .collect();

    let damages: Vec<DamageEvent> = events
        .iter()
        .filter_map(|e| match e {
            MatchEvent::Damage(d) => Some(*d),
            _ => None,
        })
        .collect();

    let total_rounds = events
        .iter()
        .filter_map(|e| match e {
            MatchEvent::RoundEnd(r) => Some(r.round),
            _ => None,
        })
        .max()
        .unwrap_or(0);

    if total_rounds == 0 {
        return vec![];
    }

    let mut roster: std::collections::BTreeSet<PlayerId> = std::collections::BTreeSet::new();
    for k in &kills {
        if let Some(attacker_id) = k.attacker_id {
            roster.insert(attacker_id);
        }
        if let Some(assister_id) = k.assister_id {
            roster.insert(assister_id);
        }
        roster.insert(k.victim_id);
    }
    for d in &damages {
        roster.insert(d.attacker_id);
        roster.insert(d.victim_id);
    }

    roster
        .into_iter()
        .map(|player_id| PlayerStats {
            player_id,
            kpr: kills_per_round(&kills, player_id, total_rounds),
            dpr: deaths_per_round(&kills, player_id, total_rounds),
            kast: kast_percentage(&kills, player_id, total_rounds),
            impact: impact_score(&kills, player_id, total_rounds),
            adr: average_damage_per_round(&damages, player_id, total_rounds),
        })
        .collect()
}

fn kills_per_round(kills: &[KillEvent], player_id: PlayerId, total_rounds: u32) -> f64 {
    let count = kills.iter().filter(|k| k.attacker_id == Some(player_id)).count();
    count as f64 / total_rounds as f64
}

fn deaths_per_round(kills: &[KillEvent], player_id: PlayerId, total_rounds: u32) -> f64 {
    let count = kills.iter().filter(|k| k.victim_id == player_id).count();
    count as f64 / total_rounds as f64
}

fn kast_percentage(kills: &[KillEvent], player_id: PlayerId, total_rounds: u32) -> f64 {
    let mut positive_rounds = 0u32;

    for round in 1..=total_rounds {
        let round_kills: Vec<&KillEvent> = kills.iter().filter(|k| k.round == round).collect();

        let got_kill = round_kills.iter().any(|k| k.attacker_id == Some(player_id));
        let got_assist = round_kills.iter().any(|k| k.assister_id == Some(player_id));
        let death = round_kills.iter().find(|k| k.victim_id == player_id);
        let survived = death.is_none();
        let traded = was_traded(&round_kills, death);

        if got_kill || got_assist || survived || traded {
            positive_rounds += 1;
        }
    }

    positive_rounds as f64 / total_rounds as f64 * 100.0
}

fn was_traded(round_kills: &[&KillEvent], death: Option<&&KillEvent>) -> bool {
    let death_event = match death {
        Some(event) => event,
        None => return false,
    };
    let killer_id = match death_event.attacker_id {
        Some(id) => id,
        None => return false,
    };
    round_kills
        .iter()
        .any(|k| k.victim_id == killer_id && k.attacker_team == Some(death_event.victim_team))
}

fn impact_score(kills: &[KillEvent], player_id: PlayerId, total_rounds: u32) -> f64 {
    let mut kills_by_round: std::collections::BTreeMap<u32, u32> = std::collections::BTreeMap::new();
    for k in kills {
        if k.attacker_id == Some(player_id) {
            *kills_by_round.entry(k.round).or_insert(0) += 1;
        }
    }
    let total_weight: u32 = kills_by_round.values().map(|count| multi_kill_weight(*count)).sum();
    total_weight as f64 / total_rounds as f64
}

fn multi_kill_weight(kills_in_round: u32) -> u32 {
    kills_in_round * (kills_in_round + 1) / 2
}

fn average_damage_per_round(damages: &[DamageEvent], player_id: PlayerId, total_rounds: u32) -> f64 {
    let total_damage: u32 = damages
        .iter()
        .filter(|d| d.attacker_id == player_id && d.attacker_team != d.victim_team)
        .map(|d| d.damage)
        .sum();
    total_damage as f64 / total_rounds as f64
}

pub fn normalize_player_stats(stats: &[PlayerStats]) -> Vec<NormalizedPlayerStats> {
    let kpr = min_max_normalize(stats.iter().map(|s| s.kpr));
    let dpr = min_max_normalize(stats.iter().map(|s| s.dpr));
    let impact = min_max_normalize(stats.iter().map(|s| s.impact));
    let adr = min_max_normalize(stats.iter().map(|s| s.adr));

    stats
        .iter()
        .enumerate()
        .map(|(i, s)| NormalizedPlayerStats {
            player_id: s.player_id,
            kpr: kpr[i],
            dpr: dpr[i],
            kast: (s.kast / 100.0).clamp(0.0, 1.0),
            impact: impact[i],
            adr: adr[i],
        })
        .collect()
}

fn min_max_normalize(values: impl Iterator<Item = f64>) -> Vec<f64> {
    let values: Vec<f64> = values.collect();
    let min = values.iter().cloned().fold(f64::INFINITY, f64::min);
    let max = values.iter().cloned().fold(f64::NEG_INFINITY, f64::max);
    let range = max - min;

    values
        .iter()
        .map(|v| if range == 0.0 { 0.5 } else { (v - min) / range })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    const TEAM_A: u8 = 2;
    const TEAM_B: u8 = 3;

    #[test]
    fn kpr_and_dpr_are_counted_per_round() {
        let p1: PlayerId = 1;
        let p2: PlayerId = 2;

        let events = vec![
            MatchEvent::Kill(KillEvent {
                round: 1,
                attacker_id: Some(p1),
                attacker_team: Some(TEAM_A),
                victim_id: p2,
                victim_team: TEAM_B,
                assister_id: None,
            }),
            MatchEvent::RoundEnd(RoundEndEvent { round: 1 }),
            MatchEvent::Kill(KillEvent {
                round: 2,
                attacker_id: Some(p2),
                attacker_team: Some(TEAM_B),
                victim_id: p1,
                victim_team: TEAM_A,
                assister_id: None,
            }),
            MatchEvent::RoundEnd(RoundEndEvent { round: 2 }),
        ];

        let stats = compute_player_stats(&events);
        let p1_stats = stats.iter().find(|s| s.player_id == p1).unwrap();
        let p2_stats = stats.iter().find(|s| s.player_id == p2).unwrap();

        assert_eq!(p1_stats.kpr, 0.5);
        assert_eq!(p1_stats.dpr, 0.5);
        assert_eq!(p2_stats.kpr, 0.5);
        assert_eq!(p2_stats.dpr, 0.5);
    }

    #[test]
    fn kast_counts_kill_assist_survive_and_trade() {
        let p1: PlayerId = 1;
        let p2: PlayerId = 2;
        let p3: PlayerId = 3;

        let events = vec![
            MatchEvent::Kill(KillEvent {
                round: 1,
                attacker_id: Some(p1),
                attacker_team: Some(TEAM_A),
                victim_id: p2,
                victim_team: TEAM_B,
                assister_id: None,
            }),
            MatchEvent::Kill(KillEvent {
                round: 1,
                attacker_id: Some(p3),
                attacker_team: Some(TEAM_B),
                victim_id: p1,
                victim_team: TEAM_A,
                assister_id: None,
            }),
            MatchEvent::RoundEnd(RoundEndEvent { round: 1 }),
        ];

        let stats = compute_player_stats(&events);
        let p2_stats = stats.iter().find(|s| s.player_id == p2).unwrap();
        let p3_stats = stats.iter().find(|s| s.player_id == p3).unwrap();

        assert_eq!(p2_stats.kast, 100.0);
        assert_eq!(p3_stats.kast, 100.0);
    }

    #[test]
    fn impact_rewards_multi_kill_rounds() {
        let p1: PlayerId = 1;
        let p2: PlayerId = 2;
        let p3: PlayerId = 3;

        let events = vec![
            MatchEvent::Kill(KillEvent {
                round: 1,
                attacker_id: Some(p1),
                attacker_team: Some(TEAM_A),
                victim_id: p2,
                victim_team: TEAM_B,
                assister_id: None,
            }),
            MatchEvent::Kill(KillEvent {
                round: 1,
                attacker_id: Some(p1),
                attacker_team: Some(TEAM_A),
                victim_id: p3,
                victim_team: TEAM_B,
                assister_id: None,
            }),
            MatchEvent::RoundEnd(RoundEndEvent { round: 1 }),
            MatchEvent::RoundEnd(RoundEndEvent { round: 2 }),
        ];

        let stats = compute_player_stats(&events);
        let p1_stats = stats.iter().find(|s| s.player_id == p1).unwrap();

        assert_eq!(p1_stats.impact, 1.5);
    }

    #[test]
    fn adr_excludes_team_damage() {
        let p1: PlayerId = 1;
        let p2: PlayerId = 2;
        let p3: PlayerId = 3;

        let events = vec![
            MatchEvent::Damage(DamageEvent {
                round: 1,
                attacker_id: p1,
                attacker_team: TEAM_A,
                victim_id: p2,
                victim_team: TEAM_B,
                damage: 40,
            }),
            MatchEvent::Damage(DamageEvent {
                round: 1,
                attacker_id: p1,
                attacker_team: TEAM_A,
                victim_id: p3,
                victim_team: TEAM_A,
                damage: 20,
            }),
            MatchEvent::RoundEnd(RoundEndEvent { round: 1 }),
            MatchEvent::RoundEnd(RoundEndEvent { round: 2 }),
        ];

        let stats = compute_player_stats(&events);
        let p1_stats = stats.iter().find(|s| s.player_id == p1).unwrap();

        assert_eq!(p1_stats.adr, 20.0);
    }

    #[test]
    fn normalize_scales_metrics_between_zero_and_one() {
        let stats = vec![
            PlayerStats { player_id: 1, kpr: 0.5, dpr: 0.5, kast: 50.0, impact: 1.0, adr: 60.0 },
            PlayerStats { player_id: 2, kpr: 1.0, dpr: 0.2, kast: 80.0, impact: 2.0, adr: 90.0 },
        ];

        let normalized = normalize_player_stats(&stats);

        assert_eq!(normalized[0].kpr, 0.0);
        assert_eq!(normalized[1].kpr, 1.0);
        assert_eq!(normalized[0].kast, 0.5);
        assert_eq!(normalized[1].kast, 0.8);
    }

    #[test]
    fn no_rounds_produces_no_stats() {
        let events: Vec<MatchEvent> = vec![];
        assert_eq!(compute_player_stats(&events), vec![]);
    }
}
