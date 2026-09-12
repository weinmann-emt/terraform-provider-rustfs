resource "rustfs_kms_key" "app" {
  name         = "app-master-key"
  enabled      = true
  skip_destroy = true
}