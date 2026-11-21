use crate::demo::DemoData;
use crate::error::BoxError;
use crate::spatial::PositionEvent;
use crate::stats::PlayerStats;
use sqlx::PgPool;

pub async fn insert_demo(pool: &PgPool, bucket: &str, key: &str, data: &DemoData) -> Result<i64, BoxError> {
    let players = serde_json::to_value(&data.players)?;
    let kills = serde_json::to_value(&data.kills)?;

    let demo_id = sqlx::query_scalar::<_, i64>(
        "INSERT INTO demos (s3_bucket, s3_key, team_t_score, team_ct_score, winner, players, kills, map_name)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
         ON CONFLICT (s3_bucket, s3_key) DO UPDATE SET
             team_t_score = EXCLUDED.team_t_score,
             team_ct_score = EXCLUDED.team_ct_score,
             winner = EXCLUDED.winner,
             players = EXCLUDED.players,
             kills = EXCLUDED.kills,
             map_name = EXCLUDED.map_name
         RETURNING id",
    )
    .bind(bucket)
    .bind(key)
    .bind(data.final_score.team_t_score)
    .bind(data.final_score.team_ct_score)
    .bind(&data.final_score.winner)
    .bind(players)
    .bind(kills)
    .bind(&data.map_name)
    .fetch_one(pool)
    .await?;

    Ok(demo_id)
}

pub async fn insert_player_stats(pool: &PgPool, demo_id: i64, stats: &[PlayerStats]) -> Result<(), BoxError> {
    for player in stats {
        sqlx::query(
            "INSERT INTO player_stats (demo_id, steamid, kpr, dpr, kast, impact, adr)
             VALUES ($1, $2, $3, $4, $5, $6, $7)
             ON CONFLICT (demo_id, steamid) DO UPDATE SET
                 kpr = EXCLUDED.kpr,
                 dpr = EXCLUDED.dpr,
                 kast = EXCLUDED.kast,
                 impact = EXCLUDED.impact,
                 adr = EXCLUDED.adr",
        )
        .bind(demo_id)
        .bind(player.player_id as i64)
        .bind(player.kpr)
        .bind(player.dpr)
        .bind(player.kast)
        .bind(player.impact)
        .bind(player.adr)
        .execute(pool)
        .await?;
    }

    Ok(())
}

pub async fn insert_positions(pool: &PgPool, demo_id: i64, positions: &[PositionEvent]) -> Result<(), BoxError> {
    for position in positions {
        sqlx::query(
            "INSERT INTO player_positions (demo_id, steamid, round, tick, team, x, y, z)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
             ON CONFLICT (demo_id, steamid, round, tick) DO NOTHING",
        )
        .bind(demo_id)
        .bind(position.player_id as i64)
        .bind(position.round as i32)
        .bind(position.tick)
        .bind(format!("{:?}", position.team).to_lowercase())
        .bind(position.position.x)
        .bind(position.position.y)
        .bind(position.position.z)
        .execute(pool)
        .await?;
    }

    Ok(())
}
