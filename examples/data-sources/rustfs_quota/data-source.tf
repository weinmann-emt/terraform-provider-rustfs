data "rustfs_quota" "test" {
  bucket = "my-bucket"
}

output "quota_bytes" {
  value = data.rustfs_quota.test.quota
}
