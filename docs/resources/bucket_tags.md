---
page_title: "rustfs_bucket_tags Resource - rustfs"
description: |-
  Manage the tags on an S3 bucket in rustfs
---

# rustfs_bucket_tags (Resource)

Manage the tags on an S3 bucket in rustfs.

## Example Usage

```terraform
resource "rustfs_bucket" "example" {
  name = "my-tagged-bucket"
}

resource "rustfs_bucket_tags" "example" {
  bucket = rustfs_bucket.example.name

  tags = {
    environment = "production"
    team        = "platform"
  }
}
```

## Schema

### Required

- `bucket` (String) Name of the bucket. Changing this forces a new resource to be created.
- `tags` (Map of String) Map of key/value tag pairs to apply to the bucket

### Read-Only

- `id` (String) The bucket name

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_tags.example my-bucket
```