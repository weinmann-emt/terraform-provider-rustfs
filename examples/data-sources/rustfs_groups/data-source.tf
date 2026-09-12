data "rustfs_groups" "all" {}

output "group_names" {
  value = data.rustfs_groups.all.groups
}