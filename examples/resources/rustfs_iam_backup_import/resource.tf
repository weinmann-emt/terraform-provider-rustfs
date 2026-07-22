data "local_file" "import_file" {
  filename = "iam-backup.zip"
}

resource "rustfs_iam_backup_import" "restore" {
  content_base64 = data.local_file.import_file.content_base64
}
