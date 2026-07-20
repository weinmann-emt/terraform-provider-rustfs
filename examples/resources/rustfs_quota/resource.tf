resource "rustfs_quota" "example" {
  bucket = rustfs_bucket.example.name
  quota  = 10737418240 # 10 GiB
}
