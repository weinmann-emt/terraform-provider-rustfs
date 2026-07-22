resource "rustfs_serviceaccount" "ci_token" {
  access_key  = "ci-bot"
  secret_key  = "s3cret-token"
  name        = "CI Pipeline"
  description = "Token for CI/CD pipeline access"
  user        = "myuser"
}
