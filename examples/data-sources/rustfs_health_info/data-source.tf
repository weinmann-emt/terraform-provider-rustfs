data "rustfs_health_info" "cluster" {}

output "rustfs_version" {
  value = data.rustfs_health_info.cluster.version
}

output "health_raw" {
  value = data.rustfs_health_info.cluster.health_info
}