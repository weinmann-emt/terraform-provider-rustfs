data "rustfs_server_info" "cluster" {}

output "cluster_mode" {
  value = data.rustfs_server_info.cluster.mode
}

output "deployment_id" {
  value = data.rustfs_server_info.cluster.deployment_id
}

output "server_versions" {
  value = data.rustfs_server_info.cluster.servers[*].version
}