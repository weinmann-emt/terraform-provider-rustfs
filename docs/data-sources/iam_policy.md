---
page_title: "rustfs_iam_policy Data Source - rustfs"
description: |-
  Inspect a single canned IAM policy
---

# rustfs_iam_policy (Data Source)

Fetch the details of a single canned IAM policy by name.

## Example Usage

```terraform
data "rustfs_iam_policy" "readwrite" {
  name = "readwrite"
}

output "readwrite_actions" {
  value = data.rustfs_iam_policy.readwrite.statement
}
```

## Schema

### Required

- `name` (String) Name of the canned policy.

### Read-Only

- `statement` (List of Object) Policy statements. Each entry has:
  - `effect` (String) Effect (Allow or Deny).
  - `action` (Set of String) Allowed or denied actions.
  - `resource` (Set of String) Resource ARNs the statement applies to.
- `version` (String) Policy version (2012-10-17).