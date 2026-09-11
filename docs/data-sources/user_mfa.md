---
page_title: "rustfs_user_mfa Data Source - rustfs"
description: |-
  Read the MFA status of a RustFS user
---

# rustfs_user_mfa (Data Source)

Read the second-factor (MFA) status of a RustFS user by access key.

The admin API only allows inspecting a user's MFA state; enrollment is performed by the user through the console.

## Example Usage

```terraform
data "rustfs_user_mfa" "alice" {
  access_key = rustfs_user.alice.access_key
}

output "alice_mfa_enabled" {
  value = data.rustfs_user_mfa.alice.enabled
}
```

## Schema

### Required

- `access_key` (String) Access key of the user.

### Read-Only

- `activated_at` (String) RFC3339 timestamp of when the second factor was activated, if enabled.
- `enabled` (Boolean) Whether two-factor authentication is enabled for the user.
- `recovery_codes_remaining` (Number) Number of unused recovery codes remaining for the user.