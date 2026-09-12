resource "rustfs_bucket" "example" {
  name = "my-bucket"
}

resource "rustfs_bucket_public_access_block" "example" {
  bucket                  = rustfs_bucket.example.name
  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = true
  restrict_public_buckets = true
}