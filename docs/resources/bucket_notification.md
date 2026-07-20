---
page_title: "rustfs_bucket_notification Resource - rustfs"
description: |-
  Manage RustFS bucket event notifications
---

# rustfs_bucket_notification (Resource)

Manage RustFS bucket event notification configuration.

## Example Usage

```terraform
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
```

## Schema

### Required

- `bucket` (String) Name of the bucket. Changing this forces a new resource.

### Optional

- `queue` (Attributes List) Queue notification configurations. Each queue has:
  - `arn` (String, Required) ARN of the queue target.
  - `events` (Set of String, Required) S3 event types.
  - `filter_prefix` (String, Optional) Filter by object key prefix.
  - `filter_suffix` (String, Optional) Filter by object key suffix.

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_notification.example my-bucket
```
