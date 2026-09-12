resource "rustfs_bucket" "example" {
  name = "my-tagged-bucket"
}

resource "rustfs_bucket_tags" "example" {
  bucket = rustfs_bucket.example.name

  tags = {
    environment = "production"
    team        = "platform"
  }
}