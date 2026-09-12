resource "rustfs_config" "primary" {
  sub_system = "notify_webhook:primary"
  settings = {
    endpoint = "http://notify.example.com/rustfs/events"
  }
}