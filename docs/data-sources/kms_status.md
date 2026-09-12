---
page_title: "rustfs_kms_status Data Source - rustfs"
description: |-
  Read the RustFS KMS status and backend configuration
---

# rustfs_kms_status (Data Source)

Read the RustFS KMS status and backend configuration from the admin endpoints
`GET /rustfs/admin/v3/kms/status` and `GET /rustfs/admin/v3/kms/config`.

The data source exposes the most relevant fields as typed attributes and the
full JSON documents verbatim via `status_json` and `config_json` for full
fidelity against the server.

> **Note:** The KMS subsystem is only populated once a backend has been
> configured and started on the server. On a stock deployment that has not
> configured KMS, reading this data source fails with `KMS service not
> initialized`.

## Example Usage

```terraform
data "rustfs_kms_status" "test" {}

output "kms_backend" {
  value = data.rustfs_kms_status.test.backend_type
}

output "kms_status_raw" {
  value = jsondecode(data.rustfs_kms_status.test.status_json)
}
```

## Schema

### Read-Only

- `backend` (String) Configured KMS backend as reported by /kms/config (e.g. local, vault-kv2, aws).
- `backend_status` (String) Health status of the KMS backend (healthy, unhealthy or error).
- `backend_type` (String) Name or type of the configured KMS backend as reported by /kms/status (e.g. local, vault-kv2, aws).
- `cache_enabled` (Boolean) Whether the KMS key cache is enabled.
- `cache_max_keys` (Number) Maximum number of keys held in the KMS cache.
- `cache_ttl_seconds` (Number) Time-to-live of cached KMS entries, in seconds.
- `config_json` (String) Full raw JSON document returned by `GET /rustfs/admin/v3/kms/config`.
- `default_key_id` (String) Key ID used when no explicit key is specified.
- `id` (String) Fixed identifier for the KMS status data source.
- `status_json` (String) Full raw JSON document returned by `GET /rustfs/admin/v3/kms/status`.

## Response Shapes

With a local KMS backend configured, `status_json` is a document of the form:

```json
{
  "backend_type": "local",
  "backend_status": "healthy",
  "cache_enabled": true,
  "cache_stats": {
    "hit_count": 0,
    "miss_count": 0,
    "entry_count": 0,
    "eviction_count": 0
  },
  "default_key_id": "rustfs-master",
  "capabilities": {
    "encrypt": true,
    "decrypt": true,
    "generate_data_key": true,
    "rotate": false,
    "enable_disable": true,
    "schedule_deletion": true,
    "versioning": false,
    "physical_delete": true,
    "update_key_metadata": true,
    "rewrap": false,
    "production_supported": false
  },
  "cluster_config": {
    "consistent": true,
    "nodes": [
      { "host": "local", "config_fingerprint": "ddac6b84ebf44e537007898038db8f282165c5c25d3be892ffd38701cf5f923d" }
    ]
  }
}
```

and `config_json` is a document of the form:

```json
{
  "backend": "local",
  "cache_enabled": true,
  "cache_max_keys": 1000,
  "cache_ttl_seconds": 900,
  "default_key_id": "rustfs-master"
}
```
