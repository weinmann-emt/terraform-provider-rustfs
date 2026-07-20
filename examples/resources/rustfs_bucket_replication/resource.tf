resource "rustfs_bucket" "source" {
  name = "source-bucket"
}

resource "rustfs_bucket_replication" "example" {
  bucket            = rustfs_bucket.source.name
  role              = "arn:minio:replication::id:source-bucket"
  destination_bucket = "arn:aws:s3:::dest-bucket"
  priority          = 1
  status            = "Enabled"
}
