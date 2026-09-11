data "rustfs_user_mfa" "alice" {
  access_key = rustfs_user.alice.access_key
}

output "alice_mfa_enabled" {
  value = data.rustfs_user_mfa.alice.enabled
}