---
page_title: "rustfs_quota Data Source - rustfs"
description: |-
  Read the quota of a RustFS bucket
---

# rustfs_quota (Data Source)

Read the quota of a RustFS bucket.

The RustFS server currently only supports the `HARD` quota type; requests for
other types (e.g. `SOFT`) are rejected with `InvalidArgument`.

## Example Usage

```terraform
data "rustfs_quota" "test" {
  bucket = "my-bucket"
}

output "quota_bytes" {
  value = data.rustfs_quota.test.quota
}
```

## Schema

### Required

- `bucket` (String) Name of the bucket.

### Read-Only

- `quota` (Number) Bytes of the quota.
- `quota_type` (String) Type of the quota, e.g. HARD. Only `HARD` is supported by the server.
