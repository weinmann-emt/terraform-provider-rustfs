---
page_title: "rustfs_bucket_replication Resource - rustfs"
description: |-
  Manage RustFS bucket replication
---

# rustfs_bucket_replication (Resource)

Manage RustFS bucket replication configuration.

## Example Usage

```terraform
resource "rustfs_bucket" "source" {
  name = "source-bucket"
}

resource "rustfs_bucket_replication" "example" {
  bucket             = rustfs_bucket.source.name
  role               = "arn:minio:replication::id:source-bucket"
  destination_bucket = "arn:aws:s3:::dest-bucket"
  priority           = 1
  status             = "Enabled"
}
```

## Schema

### Required

- `bucket` (String) Source bucket name. Changing this forces a new resource.
- `role` (String) Replication role ARN.
- `destination_bucket` (String) Destination bucket ARN.

### Optional

- `status` (String) Rule status: Enabled or Disabled. Default: Enabled.
- `priority` (Number) Rule priority. Default: 1.
- `delete_marker_replication` (String) Delete marker replication: Enabled or Disabled.
- `delete_replication` (String) Delete replication: Enabled or Disabled.

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_replication.example my-bucket
```
