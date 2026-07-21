use crate::error::BoxError;
use ahash::AHashMap;
use demoparser2::first_pass::parser_settings::{rm_user_friendly_names, ParserInputs};
use demoparser2::parse_demo::{Parser, ParsingMode};
use demoparser2::second_pass::game_events::{EventField, GameEvent};
use demoparser2::second_pass::parser_settings::create_huffman_lookup_table;
use demoparser2::second_pass::variants::Variant;
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
}

const TEAM_ROUNDS_TOTAL: &str = "team_rounds_total";

pub fn extract(demo_bytes: &[u8]) -> Result<DemoData, BoxError> {
    let huffman_lookup_table = create_huffman_lookup_table();

    let real_other_prop = rm_user_friendly_names(&vec![TEAM_ROUNDS_TOTAL.to_string()])?
        .into_iter()
        .next()
        .ok_or("failed to resolve team_rounds_total prop")?;

    let mut real_name_to_og_name = AHashMap::default();
    real_name_to_og_name.insert(real_other_prop.clone(), TEAM_ROUNDS_TOTAL.to_string());

    let settings = ParserInputs {
        real_name_to_og_name,
        wanted_players: vec![],
        wanted_player_props: vec![],
        wanted_other_props: vec![real_other_prop],
        wanted_prop_states: AHashMap::default(),
        wanted_events: vec!["player_death".to_string(), "round_end".to_string()],
        parse_ents: false,
        wanted_ticks: vec![],
        parse_projectiles: false,
        parse_grenades: false,
        only_header: false,
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

    Ok(DemoData { players, final_score, kills })
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
