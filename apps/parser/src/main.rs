use aws_config::BehaviorVersion;
use aws_sdk_s3::Client as S3Client;
use aws_sdk_sqs::Client as SqsClient;
use parser::config::Config;
use parser::error::BoxError;
use parser::{db, demo, queue, storage};
use sqlx::postgres::PgPoolOptions;
use sqlx::PgPool;

#[tokio::main]
async fn main() -> Result<(), BoxError> {
    let cfg = Config::load()?;

    let aws_cfg = aws_config::defaults(BehaviorVersion::latest())
        .region(aws_config::Region::new(cfg.aws_region.clone()))
        .load()
        .await;

    let s3_client = S3Client::new(&aws_cfg);
    let sqs_client = SqsClient::new(&aws_cfg);

    let pool = PgPoolOptions::new()
        .max_connections(5)
        .connect(&cfg.database_url)
        .await?;

    sqlx::migrate!("./migrations").run(&pool).await?;

    loop {
        let messages = queue::poll(&sqs_client, &cfg.sqs_queue_url).await?;
        for message in messages {
            if let Err(err) = process(&s3_client, &pool, &message.upload).await {
                eprintln!("failed to process {}/{}: {err}", message.upload.bucket, message.upload.key);
                continue;
            }
            if let Err(err) = queue::delete(&sqs_client, &cfg.sqs_queue_url, &message.receipt_handle).await {
                eprintln!("failed to delete message: {err}");
            }
        }
    }
}

async fn process(s3_client: &S3Client, pool: &PgPool, upload: &queue::UploadMessage) -> Result<(), BoxError> {
    let demo_bytes = storage::download(s3_client, &upload.bucket, &upload.key).await?;
    let data = demo::extract(&demo_bytes)?;
    db::insert_demo(pool, &upload.bucket, &upload.key, &data).await?;
    Ok(())
}
