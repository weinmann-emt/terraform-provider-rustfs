---
page_title: "rustfs_bucket_versioning Resource - rustfs"
description: |-
  Manage RustFS bucket versioning configuration
---

# rustfs_bucket_versioning (Resource)

Manage RustFS bucket versioning configuration.

## Example Usage

```terraform
resource "rustfs_bucket" "example" {
  name = "my-bucket"
}

resource "rustfs_bucket_versioning" "example" {
  bucket = rustfs_bucket.example.name
  status = "Enabled"
}
```

## Schema

### Required

- `bucket` (String) Name of the bucket. Changing this forces a new resource to be created.
- `status` (String) Versioning status: `Enabled` or `Suspended`.

## Import

Import is supported using the bucket name:

```
terraform import rustfs_bucket_versioning.example my-bucket
```
