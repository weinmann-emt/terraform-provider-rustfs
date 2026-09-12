---
page_title: "rustfs_server_info Data Source - rustfs"
description: |-
  Exposes cluster, server, pool and drive information from the RustFS admin API.
---

# rustfs_server_info (Data Source)

Exposes cluster, server, pool and drive information from the RustFS admin API (`GET /rustfs/admin/v3/info`).

The data source exposes the most commonly used cluster attributes as first-class
schema attributes, plus a `raw_json` attribute containing the full, unmodified
JSON response for complete fidelity.

## Example Usage

```terraform
data "rustfs_server_info" "cluster" {}

output "cluster_mode"     { value = data.rustfs_server_info.cluster.mode }
output "deployment_id"    { value = data.rustfs_server_info.cluster.deployment_id }
output "server_versions"  { value = data.rustfs_server_info.cluster.servers[*].version }
```

## Schema

### Read-Only

- `backend_type` (String) Backend type (e.g. Erasure).
- `bitrot_selftest` (String) Bitrot self test status.
- `bucket_count` (Number) Number of buckets in the cluster.
- `delete_marker_count` (Number) Number of delete markers in the cluster.
- `deployment_id` (String) Cluster deployment ID.
- `mode` (String) Cluster mode (e.g. online, offline).
- `object_count` (Number) Number of objects in the cluster.
- `offline_disks` (Number) Number of offline disks.
- `online_disks` (Number) Number of online disks.
- `pool_count` (Number) Number of storage pools in the cluster.
- `pools` (Attributes List) Storage pools and their erasure sets. See [below](#nestedatt--pools).
- `raw_json` (String) Full raw JSON response from the /info endpoint.
- `region` (String) Cluster region, when configured.
- `servers` (Attributes List) Server nodes in the cluster. See [below](#nestedatt--servers).
- `total_drives_per_set` (List of Number) Total drives per erasure set.
- `total_sets` (List of Number) Total number of erasure sets per pool.
- `usage_size` (Number) Aggregate usage size in bytes.
- `version_count` (Number) Number of object versions in the cluster.

<a id="nestedatt--pools"></a>
### Nested Schema for `pools`

Read-Only:

- `delete_markers_count` (Number)
- `heal_disks` (Number)
- `id` (Number)
- `objects_count` (Number)
- `pool_number` (Number)
- `raw_capacity` (Number)
- `raw_usage` (Number)
- `set_number` (Number)
- `usage` (Number)
- `versions_count` (Number)

<a id="nestedatt--servers"></a>
### Nested Schema for `servers`

Read-Only:

- `drives` (Attributes List) Drives attached to this server. See [below](#nestedatt--servers--drives).
- `endpoint` (String)
- `max_procs` (Number)
- `mem_alloc` (Number)
- `mem_total_alloc` (Number)
- `num_cpu` (Number)
- `state` (String)
- `uptime` (Number)
- `version` (String)

<a id="nestedatt--servers--drives"></a>
### Nested Schema for `servers.drives`

Read-Only:

- `availspace` (Number)
- `endpoint` (String)
- `healing` (Boolean)
- `local` (Boolean)
- `path` (String)
- `runtime_state` (String)
- `state` (String)
- `totalspace` (Number)
- `usedspace` (Number)
- `utilization` (Number)
- `uuid` (String)