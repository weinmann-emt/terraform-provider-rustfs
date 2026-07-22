---
page_title: "rustfs_bucket_object_lock Resource - rustfs"
description: |-
  Manage RustFS bucket object lock
---

# rustfs_bucket_object_lock (Resource)

Manage RustFS bucket object lock configuration.

## Example Usage

```terraform
resource "rustfs_bucket_object_lock" "example" {
  bucket = rustfs_bucket.example.name
  mode   = "COMPLIANCE"
  days   = 365
}
```

## Schema

### Required

- `bucket` (String) Name of the bucket. Bucket must have been created with object lock enabled.
- `mode` (String) Retention mode: `COMPLIANCE` or `GOVERNANCE`.

### Optional

- `days` (Number) Retention period in days.
- `years` (Number) Retention period in years.

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_object_lock.example my-bucket
```
