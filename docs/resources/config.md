---
page_title: "rustfs_config Resource - rustfs"
description: |-
  Manage server sub-system configuration (config-kv)
---

# rustfs_config (Resource)

Manage server sub-system configuration (config-kv), the `mc admin config` equivalent. Each resource manages one sub-system scope.

## Example Usage

```terraform
resource "rustfs_config" "primary" {
  sub_system = "notify_webhook:primary"
  settings = {
    endpoint = "http://notify.example.com/rustfs/events"
  }
}
```

## Schema

### Required

- `sub_system` (String) Sub-system scope to manage, e.g. `notify_webhook` or `notify_webhook:primary`. Changing this forces recreation.
- `settings` (Map of String) Key/value settings applied to the sub-system scope.

### Read-Only

- `id` (String) Resource identifier (the sub-system scope).

## Import

Import is supported using the sub-system scope:

```
terraform import rustfs_config.example notify_webhook:primary
```

## Limitations

- The resource manages a single sub-system scope (`subsystem` or `subsystem:target`). Multi-target sub-systems (e.g. `notify_webhook`) require an explicit `:target` suffix per resource.
- The server only renders non-default settings in `get-config-kv`. Keys whose value matches the server default (or that the server hides, such as `enable` when enabled) are not returned by Read and will show a perpetual diff. Configure only keys that round-trip through `get-config-kv`.
- `Delete` removes the whole scope (sub-system or target), resetting it to server defaults.