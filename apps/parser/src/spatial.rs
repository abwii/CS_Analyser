use crate::stats::PlayerId;

pub const DEFENSE_SETUP_OFFSET_SECONDS: f64 = 15.0;

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Vec3 {
    pub x: f64,
    pub y: f64,
    pub z: f64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Team {
    Attack,
    Defense,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct PositionEvent {
    pub round: u32,
    pub round_start_tick: i64,
    pub tick: i64,
    pub team: Team,
    pub player_id: PlayerId,
    pub position: Vec3,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct BoundingBox {
    pub min: Vec3,
    pub max: Vec3,
}

impl BoundingBox {
    pub fn contains(&self, point: &Vec3) -> bool {
        point.x >= self.min.x
            && point.x <= self.max.x
            && point.y >= self.min.y
            && point.y <= self.max.y
            && point.z >= self.min.z
            && point.z <= self.max.z
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct Zone {
    pub name: String,
    pub bounds: BoundingBox,
}

#[derive(Debug, Clone, PartialEq)]
pub struct ZonePresence {
    pub zone_name: String,
    pub percentage: f64,
}

#[derive(Debug, Clone, PartialEq)]
pub enum TacticalRole {
    Anchor(String),
    Rotator,
}

#[derive(Debug, Clone, PartialEq)]
pub struct TacticalProfile {
    pub player_id: PlayerId,
    pub role: TacticalRole,
    pub zone_presence: Vec<ZonePresence>,
}

pub fn positions_at_round_offset(
    events: &[PositionEvent],
    team: Team,
    offset_seconds: f64,
    tick_rate: f64,
) -> Vec<PositionEvent> {
    let offset_ticks = (offset_seconds * tick_rate).round() as i64;
    let mut nearest_by_round_player: std::collections::BTreeMap<(u32, PlayerId), PositionEvent> =
        std::collections::BTreeMap::new();

    for event in events {
        if event.team != team {
            continue;
        }

        let target_tick = event.round_start_tick + offset_ticks;
        let distance = (event.tick - target_tick).abs();
        let key = (event.round, event.player_id);

        match nearest_by_round_player.get(&key) {
            Some(existing) => {
                let existing_target = existing.round_start_tick + offset_ticks;
                let existing_distance = (existing.tick - existing_target).abs();
                if distance < existing_distance {
                    nearest_by_round_player.insert(key, *event);
                }
            }
            None => {
                nearest_by_round_player.insert(key, *event);
            }
        }
    }

    nearest_by_round_player.into_values().collect()
}

pub fn compute_zone_presence(events: &[PositionEvent], zones: &[Zone], player_id: PlayerId) -> Vec<ZonePresence> {
    let player_events: Vec<&PositionEvent> = events.iter().filter(|e| e.player_id == player_id).collect();
    let total = player_events.len();

    if total == 0 {
        return vec![];
    }

    zones
        .iter()
        .map(|zone| {
            let count = player_events.iter().filter(|e| zone.bounds.contains(&e.position)).count();
            ZonePresence {
                zone_name: zone.name.clone(),
                percentage: count as f64 / total as f64 * 100.0,
            }
        })
        .collect()
}

pub fn assign_tactical_roles(events: &[PositionEvent], zones: &[Zone], threshold_percentage: f64) -> Vec<TacticalProfile> {
    let mut player_ids: std::collections::BTreeSet<PlayerId> = std::collections::BTreeSet::new();
    for event in events {
        player_ids.insert(event.player_id);
    }

    player_ids
        .into_iter()
        .filter_map(|player_id| {
            let zone_presence = compute_zone_presence(events, zones, player_id);
            let dominant = zone_presence.iter().fold(None, |best: Option<&ZonePresence>, current| {
                match best {
                    Some(current_best) if current_best.percentage >= current.percentage => Some(current_best),
                    _ => Some(current),
                }
            })?;

            let role = if dominant.percentage >= threshold_percentage {
                TacticalRole::Anchor(dominant.zone_name.clone())
            } else {
                TacticalRole::Rotator
            };

            Some(TacticalProfile { player_id, role, zone_presence })
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn zone(name: &str, min: Vec3, max: Vec3) -> Zone {
        Zone { name: name.to_string(), bounds: BoundingBox { min, max } }
    }

    #[test]
    fn bounding_box_contains_checks_all_axes() {
        let bounds = BoundingBox {
            min: Vec3 { x: 0.0, y: 0.0, z: 0.0 },
            max: Vec3 { x: 10.0, y: 10.0, z: 10.0 },
        };

        assert!(bounds.contains(&Vec3 { x: 5.0, y: 5.0, z: 5.0 }));
        assert!(!bounds.contains(&Vec3 { x: 11.0, y: 5.0, z: 5.0 }));
        assert!(!bounds.contains(&Vec3 { x: 5.0, y: -1.0, z: 5.0 }));
    }

    #[test]
    fn positions_at_round_offset_picks_nearest_tick_and_filters_team() {
        let player: PlayerId = 1;

        let events = vec![
            PositionEvent {
                round: 1,
                round_start_tick: 0,
                tick: 900,
                team: Team::Defense,
                player_id: player,
                position: Vec3 { x: 1.0, y: 1.0, z: 0.0 },
            },
            PositionEvent {
                round: 1,
                round_start_tick: 0,
                tick: 960,
                team: Team::Defense,
                player_id: player,
                position: Vec3 { x: 2.0, y: 2.0, z: 0.0 },
            },
            PositionEvent {
                round: 1,
                round_start_tick: 0,
                tick: 960,
                team: Team::Attack,
                player_id: 2,
                position: Vec3 { x: 3.0, y: 3.0, z: 0.0 },
            },
        ];

        let result = positions_at_round_offset(&events, Team::Defense, DEFENSE_SETUP_OFFSET_SECONDS, 64.0);

        assert_eq!(result.len(), 1);
        assert_eq!(result[0].tick, 960);
        assert_eq!(result[0].position, Vec3 { x: 2.0, y: 2.0, z: 0.0 });
    }

    #[test]
    fn zone_presence_computes_percentage_per_zone() {
        let player: PlayerId = 1;
        let site_a = zone("A", Vec3 { x: 0.0, y: 0.0, z: 0.0 }, Vec3 { x: 10.0, y: 10.0, z: 10.0 });
        let site_b = zone("B", Vec3 { x: 100.0, y: 100.0, z: 0.0 }, Vec3 { x: 110.0, y: 110.0, z: 10.0 });

        let events = vec![
            PositionEvent { round: 1, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 5.0, y: 5.0, z: 0.0 } },
            PositionEvent { round: 2, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 5.0, y: 5.0, z: 0.0 } },
            PositionEvent { round: 3, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 105.0, y: 105.0, z: 0.0 } },
            PositionEvent { round: 4, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 5.0, y: 5.0, z: 0.0 } },
        ];

        let presence = compute_zone_presence(&events, &[site_a, site_b], player);

        assert_eq!(presence[0].zone_name, "A");
        assert_eq!(presence[0].percentage, 75.0);
        assert_eq!(presence[1].zone_name, "B");
        assert_eq!(presence[1].percentage, 25.0);
    }

    #[test]
    fn assign_tactical_roles_labels_anchor_above_threshold() {
        let player: PlayerId = 1;
        let site_a = zone("A", Vec3 { x: 0.0, y: 0.0, z: 0.0 }, Vec3 { x: 10.0, y: 10.0, z: 10.0 });
        let site_b = zone("B", Vec3 { x: 100.0, y: 100.0, z: 0.0 }, Vec3 { x: 110.0, y: 110.0, z: 10.0 });

        let events = vec![
            PositionEvent { round: 1, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 5.0, y: 5.0, z: 0.0 } },
            PositionEvent { round: 2, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 5.0, y: 5.0, z: 0.0 } },
            PositionEvent { round: 3, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 5.0, y: 5.0, z: 0.0 } },
            PositionEvent { round: 4, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 105.0, y: 105.0, z: 0.0 } },
        ];

        let profiles = assign_tactical_roles(&events, &[site_a, site_b], 70.0);

        assert_eq!(profiles.len(), 1);
        assert_eq!(profiles[0].role, TacticalRole::Anchor("A".to_string()));
    }

    #[test]
    fn assign_tactical_roles_labels_rotator_below_threshold() {
        let player: PlayerId = 1;
        let site_a = zone("A", Vec3 { x: 0.0, y: 0.0, z: 0.0 }, Vec3 { x: 10.0, y: 10.0, z: 10.0 });
        let site_b = zone("B", Vec3 { x: 100.0, y: 100.0, z: 0.0 }, Vec3 { x: 110.0, y: 110.0, z: 10.0 });

        let events = vec![
            PositionEvent { round: 1, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 5.0, y: 5.0, z: 0.0 } },
            PositionEvent { round: 2, round_start_tick: 0, tick: 0, team: Team::Defense, player_id: player, position: Vec3 { x: 105.0, y: 105.0, z: 0.0 } },
        ];

        let profiles = assign_tactical_roles(&events, &[site_a, site_b], 70.0);

        assert_eq!(profiles.len(), 1);
        assert_eq!(profiles[0].role, TacticalRole::Rotator);
    }
}
