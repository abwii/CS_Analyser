use crate::demo::DemoData;
use crate::error::BoxError;
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
