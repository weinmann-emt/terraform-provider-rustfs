resource "rustfs_policy" "readwrite" {
  name    = "readwrite"
  version = "2012-10-17"

  statement {
    effect = "Allow"
    action = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:ListBucket",
    ]
    ressource = [
      "arn:aws:s3:::my-bucket",
      "arn:aws:s3:::my-bucket/*",
    ]
  }
}
