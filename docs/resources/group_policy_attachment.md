---
page_title: "rustfs_group_policy_attachment Resource - rustfs"
description: |-
  Attach a canned IAM policy to an IAM group
---

# rustfs_group_policy_attachment (Resource)

Attach a canned IAM policy to an IAM group. The policy is detached when the resource is destroyed.

## Example Usage

```terraform
resource "rustfs_group_policy_attachment" "developers_readwrite" {
  group  = rustfs_group.developers.name
  policy = "readwrite"
}
```

## Schema

### Required

- `group` (String) Name of the IAM group. Changing this forces a new attachment.
- `policy` (String) Name of the canned policy. Changing this forces a new attachment.

### Read-Only

- `id` (String) Composite identifier in the form `group/policy`.

## Import

Import is supported using the composite `<group>/<policy>` identifier:

```
terraform import rustfs_group_policy_attachment.test developers/readwrite
```