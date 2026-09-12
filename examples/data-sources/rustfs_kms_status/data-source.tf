data "rustfs_kms_status" "test" {}

output "kms_backend" {
  value = data.rustfs_kms_status.test.backend_type
}

output "kms_status_raw" {
  value = jsondecode(data.rustfs_kms_status.test.status_json)
}
