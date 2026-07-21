use crate::error::BoxError;
use aws_sdk_s3::Client as S3Client;

pub async fn download(client: &S3Client, bucket: &str, key: &str) -> Result<Vec<u8>, BoxError> {
    let output = client.get_object().bucket(bucket).key(key).send().await?;
    let bytes = output.body.collect().await?.into_bytes();
    Ok(bytes.to_vec())
}
