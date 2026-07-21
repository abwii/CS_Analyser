use crate::demo::DemoData;
use crate::error::BoxError;
use crate::stats::PlayerStats;
use sqlx::PgPool;

pub async fn insert_demo(pool: &PgPool, bucket: &str, key: &str, data: &DemoData) -> Result<(), BoxError> {
    let players = serde_json::to_value(&data.players)?;
    let kills = serde_json::to_value(&data.kills)?;

    sqlx::query(
        "INSERT INTO demos (s3_bucket, s3_key, team_t_score, team_ct_score, winner, players, kills)
         VALUES ($1, $2, $3, $4, $5, $6, $7)
         ON CONFLICT (s3_bucket, s3_key) DO NOTHING",
    )
    .bind(bucket)
    .bind(key)
    .bind(data.final_score.team_t_score)
    .bind(data.final_score.team_ct_score)
    .bind(&data.final_score.winner)
    .bind(players)
    .bind(kills)
    .execute(pool)
    .await?;

    Ok(())
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
