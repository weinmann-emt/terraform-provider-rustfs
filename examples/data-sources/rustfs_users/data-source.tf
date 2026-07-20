# List all users
data "rustfs_users" "all" {}

output "all_user_keys" {
  value = data.rustfs_users.all.access_keys
}

# Filter users by bucket
data "rustfs_users" "bucket_users" {
  bucket = "my-bucket"
}
