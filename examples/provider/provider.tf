provider "rustfs" {
  # endpoint can also be set via RUSTFS_ENDPOINT env var
  endpoint = "127.0.0.1:9001"
  # access_key can also be set via RUSTFS_USER env var
  access_key = "admin"
  # access_secret can also be set via RUSTFS_SECRET env var
  access_secret = var.rustfs_secret
}
