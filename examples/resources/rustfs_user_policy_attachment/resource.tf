resource "rustfs_user" "alice" {
  access_key = "alice"
  secret_key = "superSecret123!"
  policy     = "readonly"
}

resource "rustfs_user_policy_attachment" "alice_readwrite" {
  user   = rustfs_user.alice.access_key
  policy = "readwrite"
}