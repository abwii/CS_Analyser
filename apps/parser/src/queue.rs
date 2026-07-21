use crate::error::BoxError;
use aws_sdk_sqs::Client as SqsClient;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct UploadMessage {
    pub bucket: String,
    pub key: String,
}

pub struct ReceivedMessage {
    pub receipt_handle: String,
    pub upload: UploadMessage,
}

pub async fn poll(client: &SqsClient, queue_url: &str) -> Result<Vec<ReceivedMessage>, BoxError> {
    let output = client
        .receive_message()
        .queue_url(queue_url)
        .max_number_of_messages(10)
        .wait_time_seconds(20)
        .send()
        .await?;

    let mut received = vec![];
    for message in output.messages.unwrap_or_default() {
        let (Some(body), Some(receipt_handle)) = (message.body, message.receipt_handle) else {
            continue;
        };
        let upload: UploadMessage = match serde_json::from_str(&body) {
            Ok(parsed) => parsed,
            Err(_) => continue,
        };
        received.push(ReceivedMessage { receipt_handle, upload });
    }
    Ok(received)
}

pub async fn delete(client: &SqsClient, queue_url: &str, receipt_handle: &str) -> Result<(), BoxError> {
    client
        .delete_message()
        .queue_url(queue_url)
        .receipt_handle(receipt_handle)
        .send()
        .await?;
    Ok(())
}
