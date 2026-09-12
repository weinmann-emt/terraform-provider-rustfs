---
page_title: "rustfs_health_info Data Source - rustfs"
description: |-
  Cluster health info and OBD diagnostics for RustFS
---

# rustfs_health_info (Data Source)

Cluster health info and OBD diagnostics for RustFS.

This data source reads the admin-plane `GET /rustfs/admin/v3/healthinfo` and
`GET /rustfs/admin/v3/obdinfo` endpoints. The full JSON payloads are exposed as
the raw string attributes `health_info` and `obd_info`. A subset of the health
status fields returned by the API (`version`, `region`, `timestamp`) and the
drive health states (`drives`) are also surfaced as typed attributes for
convenience.

## Example Usage

```terraform
data "rustfs_health_info" "cluster" {}

output "rustfs_version" {
  value = data.rustfs_health_info.cluster.version
}

output "health_raw" {
  value = data.rustfs_health_info.cluster.health_info
}
```

## Schema

### Read-Only

- `drives` (List of Object) List of drives and their health states. Each entry has:
  - `endpoint` (String) Drive endpoint.
  - `drive_path` (String) Drive path.
  - `state` (String) Drive state (for example ok, degraded, offline).
  - `total_space` (Number) Total drive space in bytes.
  - `used_space` (Number) Used drive space in bytes.
  - `available_space` (Number) Available drive space in bytes.
- `health_info` (String) Raw JSON response from the /healthinfo endpoint.
- `obd_info` (String) Raw JSON response from the /obdinfo endpoint.
- `region` (String) Region reported by the health info endpoint.
- `timestamp` (String) Timestamp reported by the health info endpoint.
- `version` (String) RustFS version reported by the health info endpoint.