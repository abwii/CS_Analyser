use crate::error::BoxError;
use crate::spatial;
use crate::stats;
use ahash::AHashMap;
use demoparser2::first_pass::parser_settings::{rm_user_friendly_names, ParserInputs};
use demoparser2::first_pass::prop_controller::{PropInfo, PLAYER_X_ID, PLAYER_Y_ID, PLAYER_Z_ID, TICK_ID};
use demoparser2::parse_demo::{DemoOutput, Parser, ParsingMode};
use demoparser2::second_pass::game_events::{EventField, GameEvent};
use demoparser2::second_pass::parser_settings::create_huffman_lookup_table;
use demoparser2::second_pass::variants::{PropColumn, VarVec, Variant};
use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct Player {
    pub steamid: u64,
    pub name: String,
    pub team_number: i32,
}

#[derive(Debug, Serialize)]
pub struct KillEvent {
    pub tick: i32,
    pub attacker_steamid: Option<u64>,
    pub attacker_name: Option<String>,
    pub victim_steamid: Option<u64>,
    pub victim_name: Option<String>,
    pub weapon: Option<String>,
    pub headshot: Option<bool>,
}

#[derive(Debug, Serialize)]
pub struct FinalScore {
    pub team_t_score: i32,
    pub team_ct_score: i32,
    pub winner: Option<String>,
}

pub struct DemoData {
    pub players: Vec<Player>,
    pub final_score: FinalScore,
    pub kills: Vec<KillEvent>,
    pub match_events: Vec<stats::MatchEvent>,
    pub positions: Vec<spatial::PositionEvent>,
}

const TEAM_ROUNDS_TOTAL: &str = "team_rounds_total";
const TEAM_NUM_TERRORIST: i64 = 2;
const TEAM_NUM_CT: i64 = 3;

pub fn extract(demo_bytes: &[u8]) -> Result<DemoData, BoxError> {
    let huffman_lookup_table = create_huffman_lookup_table();

    let real_other_prop = rm_user_friendly_names(&vec![TEAM_ROUNDS_TOTAL.to_string()])?
        .into_iter()
        .next()
        .ok_or("failed to resolve team_rounds_total prop")?;

    let player_prop_friendly_names = vec!["X".to_string(), "Y".to_string(), "Z".to_string(), "team_num".to_string()];
    let real_player_props = rm_user_friendly_names(&player_prop_friendly_names)?;

    let mut real_name_to_og_name = AHashMap::default();
    real_name_to_og_name.insert(real_other_prop.clone(), TEAM_ROUNDS_TOTAL.to_string());
    for (real_name, friendly_name) in real_player_props.iter().zip(&player_prop_friendly_names) {
        real_name_to_og_name.insert(real_name.clone(), friendly_name.clone());
    }

    let settings = ParserInputs {
        real_name_to_og_name,
        wanted_players: vec![],
        wanted_player_props: real_player_props,
        wanted_other_props: vec![real_other_prop],
        wanted_prop_states: AHashMap::default(),
        wanted_events: vec![
            "player_death".to_string(),
            "player_hurt".to_string(),
            "round_start".to_string(),
            "round_end".to_string(),
        ],
        parse_ents: true,
        wanted_ticks: vec![],
        parse_projectiles: false,
        parse_grenades: false,
        only_header: true,
        list_props: false,
        only_convars: false,
        huffman_lookup_table: &huffman_lookup_table,
        order_by_steamid: false,
        fallback_bytes: None,
    };

    let mut parser = Parser::new(settings, ParsingMode::Normal);
    let output = parser.parse_demo(demo_bytes).map_err(|e| format!("{e}"))?;

    let mut players_by_steamid: AHashMap<u64, Player> = AHashMap::default();
    for meta in &output.player_md {
        if let (Some(steamid), Some(name), Some(team_number)) =
            (meta.steamid, meta.name.clone(), meta.team_number)
        {
            if steamid != 0 {
                players_by_steamid.insert(steamid, Player { steamid, name, team_number });
            }
        }
    }
    let mut players: Vec<Player> = players_by_steamid.into_values().collect();
    players.sort_by_key(|p| p.steamid);

    let mut kills: Vec<KillEvent> = output
        .game_events
        .iter()
        .filter(|event| event.name == "player_death")
        .map(|event| KillEvent {
            tick: event.tick,
            attacker_steamid: field_u64(&event.fields, "attacker_steamid"),
            attacker_name: field_string(&event.fields, "attacker_name"),
            victim_steamid: field_u64(&event.fields, "user_steamid"),
            victim_name: field_string(&event.fields, "user_name"),
            weapon: field_string(&event.fields, "weapon"),
            headshot: field_bool(&event.fields, "headshot"),
        })
        .collect();
    kills.sort_by_key(|k| k.tick);

    let final_score = extract_final_score(&output.game_events);

    let round_starts = round_starts_from_events(&output.game_events);
    let match_events = extract_match_events(&output.game_events, &round_starts);
    let positions = extract_positions(&output, &round_starts);

    Ok(DemoData { players, final_score, kills, match_events, positions })
}

fn round_starts_from_events(game_events: &[GameEvent]) -> Vec<i32> {
    let mut ticks: Vec<i32> = game_events
        .iter()
        .filter(|event| event.name == "round_start")
        .map(|event| event.tick)
        .collect();
    ticks.sort_unstable();
    ticks
}

fn round_number_for_tick(round_starts: &[i32], tick: i32) -> u32 {
    let count = round_starts.iter().filter(|&&start| start <= tick).count() as u32;
    count.max(1)
}

fn round_start_tick_for_round(round_starts: &[i32], round: u32) -> i64 {
    round_starts.get((round - 1) as usize).copied().unwrap_or(0) as i64
}

fn extract_final_score(game_events: &[GameEvent]) -> FinalScore {
    let last_round_end = game_events
        .iter()
        .filter(|event| event.name == "round_end")
        .max_by_key(|event| event.tick);

    match last_round_end {
        Some(event) => FinalScore {
            team_t_score: field_i64(&event.fields, "t_team_rounds_total").unwrap_or(0) as i32,
            team_ct_score: field_i64(&event.fields, "ct_team_rounds_total").unwrap_or(0) as i32,
            winner: field_string(&event.fields, "winner"),
        },
        None => FinalScore { team_t_score: 0, team_ct_score: 0, winner: None },
    }
}

fn extract_match_events(game_events: &[GameEvent], round_starts: &[i32]) -> Vec<stats::MatchEvent> {
    let mut events = Vec::new();

    for event in game_events {
        match event.name.as_str() {
            "player_death" => {
                let victim_id = match field_u64(&event.fields, "user_steamid") {
                    Some(id) => id,
                    None => continue,
                };
                let victim_team = match field_i64(&event.fields, "user_team_num") {
                    Some(team) => team as u8,
                    None => continue,
                };
                let attacker_id = field_u64(&event.fields, "attacker_steamid");
                let attacker_team = field_i64(&event.fields, "attacker_team_num").map(|team| team as u8);
                let assister_id = field_u64(&event.fields, "assister_steamid");
                let round = round_number_for_tick(round_starts, event.tick);

                events.push(stats::MatchEvent::Kill(stats::KillEvent {
                    round,
                    attacker_id,
                    attacker_team,
                    victim_id,
                    victim_team,
                    assister_id,
                }));
            }
            "player_hurt" => {
                let attacker_id = match field_u64(&event.fields, "attacker_steamid") {
                    Some(id) => id,
                    None => continue,
                };
                let attacker_team = match field_i64(&event.fields, "attacker_team_num") {
                    Some(team) => team as u8,
                    None => continue,
                };
                let victim_id = match field_u64(&event.fields, "user_steamid") {
                    Some(id) => id,
                    None => continue,
                };
                let victim_team = match field_i64(&event.fields, "user_team_num") {
                    Some(team) => team as u8,
                    None => continue,
                };
                let damage = field_i64(&event.fields, "dmg_health").unwrap_or(0).max(0) as u32;
                let round = round_number_for_tick(round_starts, event.tick);

                events.push(stats::MatchEvent::Damage(stats::DamageEvent {
                    round,
                    attacker_id,
                    attacker_team,
                    victim_id,
                    victim_team,
                    damage,
                }));
            }
            "round_end" => {
                let round = round_number_for_tick(round_starts, event.tick);
                events.push(stats::MatchEvent::RoundEnd(stats::RoundEndEvent { round }));
            }
            _ => {}
        }
    }

    events
}

fn extract_positions(output: &DemoOutput, round_starts: &[i32]) -> Vec<spatial::PositionEvent> {
    let team_num_id = match find_prop_id(&output.prop_controller.prop_infos, "team_num") {
        Some(id) => id,
        None => return vec![],
    };

    let mut positions = Vec::new();

    for (steamid, columns) in &output.df_per_player {
        let tick_column = columns.get(&TICK_ID);
        let x_column = columns.get(&PLAYER_X_ID);
        let y_column = columns.get(&PLAYER_Y_ID);
        let z_column = columns.get(&PLAYER_Z_ID);
        let team_column = columns.get(&team_num_id);

        let row_count = tick_column.and_then(|column| column.data.as_ref()).map(varvec_len).unwrap_or(0);

        for index in 0..row_count {
            let tick = match column_i64(tick_column, index) {
                Some(value) => value as i32,
                None => continue,
            };
            let x = match column_f64(x_column, index) {
                Some(value) => value,
                None => continue,
            };
            let y = match column_f64(y_column, index) {
                Some(value) => value,
                None => continue,
            };
            let z = match column_f64(z_column, index) {
                Some(value) => value,
                None => continue,
            };
            let team_num = match column_i64(team_column, index) {
                Some(value) => value,
                None => continue,
            };

            let team = match team_num {
                TEAM_NUM_TERRORIST => spatial::Team::Attack,
                TEAM_NUM_CT => spatial::Team::Defense,
                _ => continue,
            };

            let round = round_number_for_tick(round_starts, tick);
            let round_start_tick = round_start_tick_for_round(round_starts, round);

            positions.push(spatial::PositionEvent {
                round,
                round_start_tick,
                tick: tick as i64,
                team,
                player_id: *steamid,
                position: spatial::Vec3 { x, y, z },
            });
        }
    }

    positions.sort_by_key(|p| (p.round, p.player_id, p.tick));
    positions
}

fn find_prop_id(prop_infos: &[PropInfo], friendly_name: &str) -> Option<u32> {
    prop_infos.iter().find(|prop| prop.prop_friendly_name == friendly_name).map(|prop| prop.id)
}

fn varvec_len(data: &VarVec) -> usize {
    match data {
        VarVec::Bool(v) => v.len(),
        VarVec::I32(v) => v.len(),
        VarVec::U32(v) => v.len(),
        VarVec::U64(v) => v.len(),
        VarVec::F32(v) => v.len(),
        VarVec::String(v) => v.len(),
        VarVec::StringVec(v) => v.len(),
        VarVec::U32Vec(v) => v.len(),
        VarVec::U64Vec(v) => v.len(),
        VarVec::XYVec(v) => v.len(),
        VarVec::XYZVec(v) => v.len(),
        VarVec::Stickers(v) => v.len(),
        VarVec::InputHistory(v) => v.len(),
    }
}

fn column_i64(column: Option<&PropColumn>, index: usize) -> Option<i64> {
    let data = column?.data.as_ref()?;
    match data {
        VarVec::I32(v) => v.get(index).copied().flatten().map(|value| value as i64),
        VarVec::U32(v) => v.get(index).copied().flatten().map(|value| value as i64),
        VarVec::U64(v) => v.get(index).copied().flatten().map(|value| value as i64),
        _ => None,
    }
}

fn column_f64(column: Option<&PropColumn>, index: usize) -> Option<f64> {
    let data = column?.data.as_ref()?;
    match data {
        VarVec::F32(v) => v.get(index).copied().flatten().map(|value| value as f64),
        VarVec::I32(v) => v.get(index).copied().flatten().map(|value| value as f64),
        _ => None,
    }
}

fn field(fields: &[EventField], name: &str) -> Option<Variant> {
    fields.iter().find(|f| f.name == name).and_then(|f| f.data.clone())
}

fn field_string(fields: &[EventField], name: &str) -> Option<String> {
    match field(fields, name)? {
        Variant::String(s) => Some(s),
        _ => None,
    }
}

fn field_bool(fields: &[EventField], name: &str) -> Option<bool> {
    match field(fields, name)? {
        Variant::Bool(b) => Some(b),
        _ => None,
    }
}

fn field_u64(fields: &[EventField], name: &str) -> Option<u64> {
    field_i64(fields, name).map(|n| n as u64)
}

fn field_i64(fields: &[EventField], name: &str) -> Option<i64> {
    match field(fields, name)? {
        Variant::I32(n) => Some(n as i64),
        Variant::U32(n) => Some(n as i64),
        Variant::U64(n) => Some(n as i64),
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use demoparser2::second_pass::collect_data::PropType;

    fn game_event(name: &str, tick: i32, fields: Vec<EventField>) -> GameEvent {
        GameEvent { name: name.to_string(), fields, tick }
    }

    fn field_entry(name: &str, data: Variant) -> EventField {
        EventField { name: name.to_string(), data: Some(data) }
    }

    #[test]
    fn round_starts_are_sorted_by_tick() {
        let events = vec![
            game_event("round_start", 500, vec![]),
            game_event("round_start", 100, vec![]),
            game_event("player_death", 200, vec![]),
            game_event("round_start", 900, vec![]),
        ];

        assert_eq!(round_starts_from_events(&events), vec![100, 500, 900]);
    }

    #[test]
    fn round_number_for_tick_is_one_indexed() {
        let round_starts = vec![100, 500, 900];

        assert_eq!(round_number_for_tick(&round_starts, 50), 1);
        assert_eq!(round_number_for_tick(&round_starts, 100), 1);
        assert_eq!(round_number_for_tick(&round_starts, 499), 1);
        assert_eq!(round_number_for_tick(&round_starts, 500), 2);
        assert_eq!(round_number_for_tick(&round_starts, 950), 3);
    }

    #[test]
    fn round_number_for_tick_falls_back_to_one_when_no_round_starts() {
        assert_eq!(round_number_for_tick(&[], 12345), 1);
    }

    #[test]
    fn round_start_tick_for_round_looks_up_by_index() {
        let round_starts = vec![100, 500, 900];

        assert_eq!(round_start_tick_for_round(&round_starts, 1), 100);
        assert_eq!(round_start_tick_for_round(&round_starts, 3), 900);
        assert_eq!(round_start_tick_for_round(&round_starts, 4), 0);
    }

    #[test]
    fn extract_final_score_reads_last_round_end() {
        let events = vec![
            game_event(
                "round_end",
                100,
                vec![
                    field_entry("t_team_rounds_total", Variant::U32(3)),
                    field_entry("ct_team_rounds_total", Variant::U32(1)),
                    field_entry("winner", Variant::String("TERRORIST".to_string())),
                ],
            ),
            game_event(
                "round_end",
                200,
                vec![
                    field_entry("t_team_rounds_total", Variant::U32(3)),
                    field_entry("ct_team_rounds_total", Variant::U32(2)),
                    field_entry("winner", Variant::String("CT".to_string())),
                ],
            ),
        ];

        let score = extract_final_score(&events);
        assert_eq!(score.team_t_score, 3);
        assert_eq!(score.team_ct_score, 2);
        assert_eq!(score.winner, Some("CT".to_string()));
    }

    #[test]
    fn extract_final_score_defaults_when_no_round_end() {
        let score = extract_final_score(&[]);
        assert_eq!(score.team_t_score, 0);
        assert_eq!(score.team_ct_score, 0);
        assert_eq!(score.winner, None);
    }

    #[test]
    fn extract_match_events_builds_kill_with_round_and_teams() {
        let round_starts = vec![0];
        let events = vec![game_event(
            "player_death",
            64,
            vec![
                field_entry("attacker_steamid", Variant::U64(111)),
                field_entry("attacker_team_num", Variant::I32(2)),
                field_entry("user_steamid", Variant::U64(222)),
                field_entry("user_team_num", Variant::I32(3)),
                field_entry("assister_steamid", Variant::U64(333)),
            ],
        )];

        let match_events = extract_match_events(&events, &round_starts);
        assert_eq!(match_events.len(), 1);

        match &match_events[0] {
            stats::MatchEvent::Kill(kill) => {
                assert_eq!(kill.round, 1);
                assert_eq!(kill.attacker_id, Some(111));
                assert_eq!(kill.attacker_team, Some(2));
                assert_eq!(kill.victim_id, 222);
                assert_eq!(kill.victim_team, 3);
                assert_eq!(kill.assister_id, Some(333));
            }
            _ => panic!("expected a kill event"),
        }
    }

    #[test]
    fn extract_match_events_skips_kill_missing_victim_team() {
        let events = vec![game_event(
            "player_death",
            64,
            vec![field_entry("user_steamid", Variant::U64(222))],
        )];

        assert_eq!(extract_match_events(&events, &[0]).len(), 0);
    }

    #[test]
    fn extract_match_events_builds_damage_and_clamps_missing_amount() {
        let round_starts = vec![0];
        let events = vec![game_event(
            "player_hurt",
            64,
            vec![
                field_entry("attacker_steamid", Variant::U64(111)),
                field_entry("attacker_team_num", Variant::I32(2)),
                field_entry("user_steamid", Variant::U64(222)),
                field_entry("user_team_num", Variant::I32(3)),
                field_entry("dmg_health", Variant::I32(27)),
            ],
        )];

        let match_events = extract_match_events(&events, &round_starts);
        match &match_events[0] {
            stats::MatchEvent::Damage(damage) => {
                assert_eq!(damage.round, 1);
                assert_eq!(damage.attacker_id, 111);
                assert_eq!(damage.attacker_team, 2);
                assert_eq!(damage.victim_id, 222);
                assert_eq!(damage.victim_team, 3);
                assert_eq!(damage.damage, 27);
            }
            _ => panic!("expected a damage event"),
        }
    }

    #[test]
    fn extract_match_events_builds_round_end() {
        let round_starts = vec![0, 1000];
        let events = vec![game_event("round_end", 1500, vec![])];

        let match_events = extract_match_events(&events, &round_starts);
        match &match_events[0] {
            stats::MatchEvent::RoundEnd(round_end) => assert_eq!(round_end.round, 2),
            _ => panic!("expected a round end event"),
        }
    }

    fn prop_info(id: u32, friendly_name: &str) -> PropInfo {
        PropInfo {
            id,
            prop_type: PropType::Player,
            prop_name: friendly_name.to_string(),
            prop_friendly_name: friendly_name.to_string(),
            is_player_prop: true,
        }
    }

    #[test]
    fn find_prop_id_matches_by_friendly_name() {
        let prop_infos = vec![prop_info(10, "team_num"), prop_info(11, "health")];
        assert_eq!(find_prop_id(&prop_infos, "team_num"), Some(10));
        assert_eq!(find_prop_id(&prop_infos, "missing"), None);
    }

    #[test]
    fn varvec_len_reports_the_right_length() {
        assert_eq!(varvec_len(&VarVec::I32(vec![Some(1), Some(2), None])), 3);
        assert_eq!(varvec_len(&VarVec::F32(vec![Some(1.0)])), 1);
    }

    #[test]
    fn column_i64_reads_supported_variants_and_skips_missing() {
        let column = PropColumn { data: Some(VarVec::I32(vec![Some(2), None, Some(3)])), num_nones: 1 };
        assert_eq!(column_i64(Some(&column), 0), Some(2));
        assert_eq!(column_i64(Some(&column), 1), None);
        assert_eq!(column_i64(Some(&column), 2), Some(3));
        assert_eq!(column_i64(None, 0), None);
    }

    #[test]
    fn column_f64_reads_f32_and_i32_variants() {
        let float_column = PropColumn { data: Some(VarVec::F32(vec![Some(1.5)])), num_nones: 0 };
        assert_eq!(column_f64(Some(&float_column), 0), Some(1.5));

        let int_column = PropColumn { data: Some(VarVec::I32(vec![Some(4)])), num_nones: 0 };
        assert_eq!(column_f64(Some(&int_column), 0), Some(4.0));

        let string_column = PropColumn { data: Some(VarVec::String(vec![Some("a".to_string())])), num_nones: 0 };
        assert_eq!(column_f64(Some(&string_column), 0), None);
    }
}
