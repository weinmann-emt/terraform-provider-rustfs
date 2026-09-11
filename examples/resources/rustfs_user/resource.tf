resource "rustfs_user" "example" {
  access_key  = "myuser"
  secret_key  = "supersecret"
  status      = "enabled"
  policy      = "readwrite"
}

# Rotate the secret key in place (no resource replacement):
#   secret_key = "newsupersecret"
#
# Enable or disable the user in place:
#   status = "disabled"
