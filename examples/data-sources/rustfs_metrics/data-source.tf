data "rustfs_metrics" "test" {}

output "metrics" {
  value = data.rustfs_metrics.test.metrics
}