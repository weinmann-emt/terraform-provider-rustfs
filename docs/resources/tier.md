---
page_title: "rustfs_tier Resource - rustfs"
description: |-
  Manage RustFS storage tiers
---

# rustfs_tier (Resource)

Manage RustFS storage tiers for data transition to external backends (S3, Azure, GCS, MinIO, etc.).

## Example Usage

```terraform
resource "rustfs_tier" "s3_cold" {
  name      = "S3COLD"
  tier_type = "s3"
  config_json = jsonencode({
    s3 = {
      name       = "S3COLD"
      endpoint   = "https://s3.amazonaws.com"
      access_key = "AKIAIOSFODNN7EXAMPLE"
      secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      region     = "us-east-1"
      bucket     = "cold-storage-backup"
    }
  })
}
```

## Schema

### Required

- `name` (String) Tier name (must be uppercase). Changing this forces recreation.
- `tier_type` (String) Tier type: s3, minio, azure, gcs, aliyun, tencent, huaweicloud, r2, or rustfs.
- `config_json` (String, Sensitive) Backend-specific configuration as a JSON string.

## Import

Import is supported using the tier name:

```
terraform import rustfs_tier.example MYTIER
```
