---
page_title: "rustfs_iam_backup_import Resource - rustfs"
description: |-
  Import RustFS IAM data
---

# rustfs_iam_backup_import (Resource)

Import RustFS IAM entities from a base64-encoded ZIP archive.

## Example Usage

```terraform
resource "rustfs_iam_backup_import" "restore" {
  content_base64 = data.rustfs_iam_backup.export.content_base64
}
```

## Schema

### Required

- `content_base64` (String, Sensitive) Base64-encoded ZIP archive. Changing this forces recreation.
