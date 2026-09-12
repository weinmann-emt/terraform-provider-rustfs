resource "rustfs_bucket" "example" {
  name = "example"
}

resource "rustfs_bucket_policy" "public_read" {
  bucket = rustfs_bucket.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "PublicRead"
      Effect    = "Allow"
      Principal = { AWS = ["*"] }
      Action    = ["s3:GetObject"]
      Resource  = ["arn:aws:s3:::example/*"]
    }]
  })
}
