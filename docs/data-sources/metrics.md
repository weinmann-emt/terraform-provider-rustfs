---
page_title: "rustfs_metrics Data Source - rustfs"
description: |-
  Metrics stream for RustFS server
---

# rustfs_metrics (Data Source)

Metrics stream for RustFS server

## Example Usage

```terraform
data "rustfs_metrics" "test" {}

output "metrics" {
  value = data.rustfs_metrics.test.metrics
}
```

## Schema

### Read-Only

- `metrics` (String) Raw metrics stream from the /metrics endpoint