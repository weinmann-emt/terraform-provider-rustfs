---
page_title: "rustfs_iam_backup Data Source - rustfs"
description: |-
  Export RustFS IAM data
---

# rustfs_iam_backup (Data Source)

Export all RustFS IAM entities (users, groups, policies, service accounts) as a base64-encoded ZIP archive.

## Example Usage

```terraform
data "rustfs_iam_backup" "export" {}
resource "local_file" "backup" {
  content_base64 = data.rustfs_iam_backup.export.content_base64
  filename       = "iam-backup.zip"
}
```

## Schema

### Read-Only

- `content_base64` (String, Sensitive) Base64-encoded ZIP archive.
