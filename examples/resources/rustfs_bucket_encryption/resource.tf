resource "rustfs_bucket" "encrypted" {
  name = "my-encrypted-bucket"
}

resource "rustfs_bucket_encryption" "example" {
  bucket    = rustfs_bucket.encrypted.name
  algorithm = "AES256"
}
