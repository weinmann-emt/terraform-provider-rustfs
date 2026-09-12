---
page_title: "rustfs_bucket_public_access_block Resource - rustfs"
description: |-
  Manage the block public access configuration of an S3 bucket in rustfs
---

# rustfs_bucket_public_access_block (Resource)

Manage the block public access configuration of an S3 bucket in rustfs.

## Example Usage

```terraform
resource "rustfs_bucket" "example" {
  name = "my-bucket"
}

resource "rustfs_bucket_public_access_block" "example" {
  bucket                  = rustfs_bucket.example.name
  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = true
  restrict_public_buckets = true
}
```

## Schema

### Required

- `bucket` (String) Name of the bucket. Changing this forces a new resource to be created.

### Optional

- `block_public_acls` (Boolean) Whether Amazon S3 should block public ACLs for this bucket.
- `block_public_policy` (Boolean) Whether Amazon S3 should block public bucket policies for this bucket.
- `ignore_public_acls` (Boolean) Whether Amazon S3 should ignore public ACLs for this bucket.
- `restrict_public_buckets` (Boolean) Whether Amazon S3 should restrict public bucket policies for this bucket.

### Read-Only

- `id` (String) The bucket name.

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_public_access_block.example my-bucket
```