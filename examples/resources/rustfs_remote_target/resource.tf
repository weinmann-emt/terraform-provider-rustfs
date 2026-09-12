resource "rustfs_bucket" "source" {
  name = "source-bucket"
}

resource "rustfs_bucket_versioning" "source" {
  bucket = rustfs_bucket.source.name
  status = "Enabled"
}

resource "rustfs_remote_target" "peer" {
  type          = "replication"
  endpoint      = "https://peer.example.com"
  access_key    = "peeruser"
  secret_key    = var.peer_secret
  secure        = true
  bucket        = rustfs_bucket.source.name
  target_bucket = "backup"
  depends_on    = [rustfs_bucket_versioning.source]
}
