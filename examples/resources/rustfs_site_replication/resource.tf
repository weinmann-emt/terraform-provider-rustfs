resource "rustfs_site_replication" "peer" {
  name       = "peer"
  endpoint   = "http://peer.example.com:9001"
  access_key = "rustfsadmin"
  secret_key = var.peer_secret
}