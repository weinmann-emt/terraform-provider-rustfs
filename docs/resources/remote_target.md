---
page_title: "rustfs_remote_target Resource - rustfs"
description: |-
  Manage remote targets used as replication destinations or notification ARNs
---

# rustfs_remote_target (Resource)

Manage remote targets used as replication destinations or notification ARNs in rustfs.

The remote target is registered on a source bucket. The server assigns the ARN.
`set-remote-target` requires the remote endpoint to be reachable from the rustfs
server and the destination bucket to exist on it, so the remote must be
operational before this resource can be created.

## Example Usage

```terraform
resource "rustfs_bucket" "source" {
  name = "source-bucket"
}

resource "rustfs_bucket_versioning" "source" {
  bucket = rustfs_bucket.source.name
  status = "Enabled"
}

resource "rustfs_remote_target" "peer" {
  type          = "replication"
  endpoint      = "https://peer.example.com"
  access_key    = "peeruser"
  secret_key    = "peersecret"
  secure        = true
  bucket        = rustfs_bucket.source.name
  target_bucket = "backup"
  depends_on    = [rustfs_bucket_versioning.source]
}
```

## Schema

### Required

- `access_key` (String, Sensitive) Access key for the remote target.
- `bucket` (String) Source bucket on which the remote target is registered. Changing this forces a new resource to be created.
- `endpoint` (String) Endpoint of the remote target, reachable from the rustfs server.
- `secret_key` (String, Sensitive) Secret key for the remote target.
- `target_bucket` (String) Destination bucket on the remote target. Changing this forces a new resource to be created.

### Optional

- `path` (String) Path prefix on the remote target.
- `region` (String) Region of the remote target.
- `secure` (Boolean) Use TLS for the remote target connection. Defaults to `true`.
- `type` (String) Remote target type. Only `replication` is supported by this server version. Defaults to `replication`.

### Read-Only

- `arn` (String) ARN assigned by the server to the remote target.

## Import

Import is supported using the source bucket and ARN, separated by a colon:

```
terraform import rustfs_remote_target.my_target source-bucket:arn:minio:replication::d05cb765-0e40-4f59-b7d5-c88ca51cd0b4:backup
```
