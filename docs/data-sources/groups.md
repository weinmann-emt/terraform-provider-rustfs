---
page_title: "rustfs_groups Data Source - rustfs"
description: |-
  Lists all IAM groups for RustFS
---

# rustfs_groups (Data Source)

Lists all IAM groups for RustFS.

## Example Usage

```terraform
data "rustfs_groups" "all" {}

output "group_names" {
  value = data.rustfs_groups.all.groups
}
```

## Schema

### Read-Only

- `groups` (Set of String) Set of IAM group names