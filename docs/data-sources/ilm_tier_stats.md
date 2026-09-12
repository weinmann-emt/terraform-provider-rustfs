---
page_title: "rustfs_ilm_tier_stats Data Source - rustfs"
description: |-
  Expose per-tier ILM storage statistics
---

# rustfs_ilm_tier_stats (Data Source)

Exposes per-tier ILM storage statistics (object counts and sizes) from RustFS.

## Example Usage

```terraform
data "rustfs_ilm_tier_stats" "all" {}

output "tiers" {
  value = data.rustfs_ilm_tier_stats.all.tiers
}
```

## Schema

### Read-Only

- `id` (String) Data source identifier.
- `tiers` (Attributes List) Per-tier statistics. See [below for nested schema](#nestedatt--tiers).

<a id="nestedatt--tiers"></a>
### Nested Schema for `tiers`

Read-Only:

- `name` (String) Tier name.
- `num_objects` (Number) Number of objects on the tier.
- `num_versions` (Number) Number of object versions on the tier.
- `total_size` (Number) Total object size in bytes.
