resource "rustfs_ldap_service_account" "alice_token" {
  access_key = "ldap-sa-alice"
  secret_key = "a-very-secret-key"
  name       = "alice-ci-token"
  user       = "uid=alice,ou=people,dc=example,dc=com"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject"]
        Resource = ["arn:aws:s3:::*"]
      }
    ]
  })
}