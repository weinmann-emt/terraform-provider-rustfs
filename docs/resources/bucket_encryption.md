---
page_title: "rustfs_bucket_encryption Resource - rustfs"
description: |-
  Manage RustFS bucket encryption
---

# rustfs_bucket_encryption (Resource)

Manage RustFS bucket server-side encryption configuration.

## Example Usage

```terraform
resource "rustfs_bucket" "encrypted" {
  name = "my-encrypted-bucket"
}

resource "rustfs_bucket_encryption" "example" {
  bucket    = rustfs_bucket.encrypted.name
  algorithm = "AES256"
}
```

## Schema

### Required

- `bucket` (String) Name of the bucket. Changing this forces a new resource to be created.
- `algorithm` (String) Encryption algorithm: `AES256` (SSE-S3) or `aws:kms` (SSE-KMS).

### Optional

- `kms_master_key_id` (String) KMS Master Key ID. Required when algorithm is `aws:kms`.

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_encryption.example my-bucket
```
