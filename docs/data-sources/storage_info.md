---
page_title: "rustfs_storage_info Data Source - rustfs"
description: |-
  Storage info for the RustFS cluster
---

# rustfs_storage_info (Data Source)

Storage layout and health information for the RustFS cluster, including the storage backend and a per-drive breakdown.

## Example Usage

```terraform
data "rustfs_storage_info" "cluster" {}

output "backend_type" {
  value = data.rustfs_storage_info.cluster.backend.backend_type
}

output "disk_count" {
  value = length(data.rustfs_storage_info.cluster.disks)
}

output "raw" {
  value = data.rustfs_storage_info.cluster.raw_json
}
```

## Schema

### Read-Only

- `backend` (Object) Storage backend configuration.
  - `backend_type` (String) Storage backend type (e.g. Erasure).
  - `drives_per_set` (List of Number) Number of drives per erasure set.
  - `rrsc_data` (List of Number) Reduced redundancy storage class data drives per set.
  - `rrsc_parities` (List of Number) Reduced redundancy storage class parity drives per set.
  - `standard_sc_data` (List of Number) Standard storage class data drives per set.
  - `standard_sc_parities` (List of Number) Standard storage class parity drives per set.
  - `total_sets` (List of Number) Total number of erasure sets.
- `disks` (List of Object) Per-drive storage breakdown.
  - `avail_space` (Number) Available space in bytes.
  - `disk_index` (Number) Index of the disk within the set.
  - `endpoint` (String) Disk endpoint.
  - `free_inodes` (Number) Free inodes.
  - `healing` (Boolean) Whether the drive is healing.
  - `local` (Boolean) Whether the drive is local.
  - `path` (String) Filesystem path of the drive.
  - `pool_index` (Number) Pool the drive belongs to.
  - `runtime_state` (String) Runtime state of the drive.
  - `scanning` (Boolean) Whether the drive is being scanned.
  - `set_index` (Number) Erasure set the drive belongs to.
  - `state` (String) Health state of the drive.
  - `total_space` (Number) Total space in bytes.
  - `used_inodes` (Number) Used inodes.
  - `used_space` (Number) Used space in bytes.
  - `uuid` (String) Unique identifier of the drive.
- `raw_json` (String) Raw JSON response from the /storageinfo endpoint for full fidelity.
