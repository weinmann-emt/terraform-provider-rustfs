data "rustfs_replication_metrics" "all" {
  bucket = "my-bucket"
}

output "replication_targets" {
  value = data.rustfs_replication_metrics.all.targets
}