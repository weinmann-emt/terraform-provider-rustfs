resource "rustfs_bucket" "example" {
  name = "example"
}

resource "rustfs_bucket_cors" "example" {
  bucket = rustfs_bucket.example.name

  rule {
    id              = "webapp"
    allowed_origins = ["https://app.example.com"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE"]
    allowed_headers = ["*"]
    max_age_seconds = 3000
  }
}