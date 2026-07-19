resource "rustfs_bucket_object_lock" "example" {
  bucket = rustfs_bucket.example.name
  mode   = "COMPLIANCE"
  days   = 365
}
