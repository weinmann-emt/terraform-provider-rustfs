data "rustfs_pools" "all" {}

output "pool_names" {
  value = data.rustfs_pools.all.names
}
