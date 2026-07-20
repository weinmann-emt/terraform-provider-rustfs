---
page_title: "rustfs_pools Data Source - rustfs"
description: |-
  List RustFS storage pools
---

# rustfs_pools (Data Source)

List all RustFS storage pools.

## Example Usage

```terraform
data "rustfs_pools" "all" {}
output "pool_names" { value = data.rustfs_pools.all.names }
```

## Schema

### Read-Only

- `names` (List of String) List of storage pool names.
