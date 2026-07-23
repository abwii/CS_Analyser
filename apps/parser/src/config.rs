use std::env;

pub struct Config {
    pub aws_region: String,
    pub sqs_queue_url: String,
    pub database_url: String,
    pub s3_force_path_style: bool,
}

impl Config {
    pub fn load() -> Result<Self, String> {
        dotenvy::dotenv().ok();
        Ok(Self {
            aws_region: require_env("AWS_REGION")?,
            sqs_queue_url: require_env("SQS_QUEUE_URL")?,
            database_url: require_env("DATABASE_URL")?,
            s3_force_path_style: env::var("S3_FORCE_PATH_STYLE").map(|v| v == "true").unwrap_or(false),
        })
    }
}

fn require_env(key: &str) -> Result<String, String> {
    env::var(key).map_err(|_| format!("{key} is required"))
}
