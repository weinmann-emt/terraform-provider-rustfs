resource "rustfs_user" "example" {
  access_key  = "myuser"
  secret_key  = "supersecret"
  status      = "enabled"
  policy      = "readwrite"
}
