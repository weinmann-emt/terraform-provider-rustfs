resource "rustfs_bucket" "example" {
  name = "my-versioned-bucket"
}

resource "rustfs_bucket_versioning" "example" {
  bucket = rustfs_bucket.example.name
  status = "Enabled"
}
