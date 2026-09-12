---
page_title: "rustfs_bucket_policy Resource - rustfs"
description: |-
  Manage the raw S3 bucket policy document in rustfs
---

# rustfs_bucket_policy (Resource)

Manage the raw S3 bucket policy document in rustfs.

This resource manages a complete, raw S3 bucket policy document. For canned
(readonly/writeonly/readwrite) policies use `rustfs_policy` instead.

## Example Usage

```terraform
resource "rustfs_bucket" "example" {
  name = "example"
}

resource "rustfs_bucket_policy" "public_read" {
  bucket = rustfs_bucket.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "PublicRead"
      Effect    = "Allow"
      Principal = { AWS = ["*"] }
      Action    = ["s3:GetObject"]
      Resource  = ["arn:aws:s3:::example/*"]
    }]
  })
}
```

## Schema

### Required

- `bucket` (String) Name of the bucket. Changing this forces a new resource to be created.
- `policy` (String) Raw S3 bucket policy document as JSON.

### Read-Only

- `id` (String) The bucket name.

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_policy.example example
```
