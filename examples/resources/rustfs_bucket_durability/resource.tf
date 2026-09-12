resource "rustfs_bucket" "example" {
  name = "my-bucket"
}

resource "rustfs_bucket_durability" "example" {
  bucket     = rustfs_bucket.example.name
  mode       = "strict"
  depends_on = [rustfs_bucket.example]
}