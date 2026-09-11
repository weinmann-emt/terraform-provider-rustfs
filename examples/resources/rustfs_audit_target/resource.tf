resource "rustfs_audit_target" "test" {
  target_type = "audit_webhook"
  target_name = "terraform"
  endpoint    = "https://hooks.example.com/webhook/terraform"
  auth_token  = "superSecret"
  comment     = "managed by terraform"
}