data "rustfs_ilm_tier_stats" "all" {}

output "tiers" {
  value = data.rustfs_ilm_tier_stats.all.tiers
}
