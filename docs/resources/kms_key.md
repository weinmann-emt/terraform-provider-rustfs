---
page_title: "rustfs_kms_key Resource - rustfs"
subcategory: ""
description: |-
  Manage RustFS KMS master keys
---

# rustfs_kms_key (Resource)

Manage RustFS KMS master keys: create, describe, enable/disable, rotate and
(safely) delete. The resource covers the master-key lifecycle exposed by the
RustFS admin API `kms/keys` endpoints.

> **Warning:** Deleting a KMS master key is **critical and irreversible**.
> RustFS schedules the destruction of the key material behind a pending window
> (default 30 days) and existing data encrypted under the key becomes
> permanently undecryptable once the deletion completes. To keep a key around
> after removing the resource, set `skip_destroy = true`.

## Example Usage

```terraform
resource "rustfs_kms_key" "app" {
  name         = "app-master-key"
  enabled      = true
  skip_destroy = true
}
```

## Schema

### Required

- `name` (String) Name of the KMS master key. Changing this forces recreation.

### Optional

- `enabled` (Boolean) Whether the KMS master key is enabled. Defaults to `true`.
- `skip_destroy` (Boolean) When `true`, destroy only removes Terraform state and
  leaves the irreversible server-side key deletion unexecuted. Defaults to
  `false`.

### Read-Only

- `created_at` (String) Timestamp when the key was created.
- `key_id` (String) Server-generated id of the KMS master key. On the Local dev
  backend this equals the key name.

## Import

Import is supported using the key name:

```
terraform import rustfs_kms_key.app app-master-key
```

## Backend support

The RustFS Local development backend supports key creation, describe,
enable/disable and scheduled deletion, but **not rotation** — the server answers
a rotate request with `501 Not Implemented` (capability `rotate: false`). The
Static (read-only) backend rejects key creation entirely.

- The acceptance test (`TestAccKmsKeyResource`) probes the running backend and
  skips gracefully when the backend cannot create keys.
- To exercise key rotation, run the provider against a production KMS backend
  (Vault or AWS KMS) that advertises `rotate: true`.