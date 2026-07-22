---
page_title: "rustfs_users Data Source - rustfs"
description: |-
  List RustFS IAM users
---

# rustfs_users (Data Source)

List all RustFS IAM users, optionally filtered by bucket name.

## Example Usage

```terraform
data "rustfs_users" "all" {}
output "all_users" { value = data.rustfs_users.all.access_keys }
```

## Schema

### Optional

- `bucket` (String) Filter users by bucket name.

### Read-Only

- `access_keys` (List of String) List of user access keys.
