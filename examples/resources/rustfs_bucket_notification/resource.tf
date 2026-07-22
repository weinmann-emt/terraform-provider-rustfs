resource "rustfs_bucket" "events" {
  name = "my-event-bucket"
}

resource "rustfs_bucket_notification" "example" {
  bucket = rustfs_bucket.events.name

  queue {
    arn    = "arn:minio:sqs::PRIMARY:amqp"
    events = ["s3:ObjectCreated:*", "s3:ObjectRemoved:*"]
    filter_prefix = "uploads/"
    filter_suffix = ".jpg"
  }
}
