resource "rustfs_group" "developers" {
  name    = "developers"
  status  = "enabled"
  members = [
    "alice",
    "bob",
  ]
}
