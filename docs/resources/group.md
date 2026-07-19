---
page_title: "rustfs_group Resource - rustfs"
description: |-
  Manage RustFS IAM groups
---

# rustfs_group (Resource)

Manage RustFS IAM groups and their members.

## Example Usage

```terraform
resource "rustfs_group" "developers" {
  name    = "developers"
  status  = "enabled"
  members = ["user1", "user2"]
}
```

## Schema

### Required

- `name` (String) Group name. Changing this forces a new resource to be created.

### Optional

- `members` (Set of String) User access keys that are members of this group.
- `status` (String) Group status: `enabled` or `disabled`. Defaults to `enabled`.

## Import

Import is supported using the group name:

```
terraform import rustfs_group.my_group my-group-name
```
