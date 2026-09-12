---
page_title: "rustfs_module_switch Resource - rustfs"
description: |-
  Manage RustFS feature module switches
---

# rustfs_module_switch (Resource)

Manage RustFS feature module switches (the notify and audit modules).

The server exposes two writable module switches — `notify_enabled` and
`audit_enabled` — together with read-only state describing the persisted
values and their source. This resource writes the full set of switches on
every apply. Destroying this resource leaves the server switches unchanged
(the admin API has no DELETE/reset endpoint for module switches).

The effective switch state may be overridden by the `RUSTFS_NOTIFY_ENABLE` /
`RUSTFS_AUDIT_ENABLE` environment variables on the server. When such an
override is present, the server rejects conflicting writes; update the
environment value first, then apply the switch state.

## Example Usage

```terraform
resource "rustfs_module_switch" "test" {
  notify_enabled = true
  audit_enabled  = true
}
```

## Schema

### Optional

- `audit_enabled` (Boolean) Whether the audit module is enabled. Defaults to `false`.
- `notify_enabled` (Boolean) Whether the notify module is enabled. Defaults to `true`.

### Read-Only

- `id` (String) Fixed identifier for the module switch set.

## Import

Import is supported using the fixed id `module-switches`:

```
terraform import rustfs_module_switch.test module-switches
```
