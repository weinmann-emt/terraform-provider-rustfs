data "rustfs_iam_backup" "export" {}

resource "local_file" "backup" {
  content_base64 = data.rustfs_iam_backup.export.content_base64
  filename       = "iam-backup.zip"
}
