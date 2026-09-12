resource "rustfs_bucket" "example" {
  name = "my-lifecycle-bucket"
}

resource "rustfs_bucket_lifecycle_configuration" "example" {
  bucket = rustfs_bucket.example.name

  rule {
    id     = "expire-logs"
    status = "Enabled"

    filter {
      prefix = "logs/"
    }

    expiration {
      days = 30
    }
  }

  rule {
    id     = "archive-old"
    status = "Enabled"

    filter {
      prefix = "archive/"
    }

    transition {
      days          = 60
      storage_class = "WARM"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "WARM"
    }

    noncurrent_version_expiration {
      noncurrent_days = 365
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
