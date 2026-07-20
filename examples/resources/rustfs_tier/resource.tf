resource "rustfs_tier" "s3_cold" {
  name      = "S3COLD"
  tier_type = "s3"
  config_json = jsonencode({
    s3 = {
      name       = "S3COLD"
      endpoint   = "https://s3.amazonaws.com"
      access_key = "AKIAIOSFODNN7EXAMPLE"
      secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      region     = "us-east-1"
      bucket     = "cold-storage-backup"
    }
  })
}
