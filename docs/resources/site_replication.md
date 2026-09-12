---
page_title: "rustfs_site_replication Resource - rustfs"
description: |-
  Manage site-replication peers in rustfs
---

# rustfs_site_replication (Resource)

Manage site-replication peers in rustfs.

~> **Multi-site requirement:** RustFS site replication requires a reachable
   peer site. A single-node server cannot form a site-replication group, so
   this resource can only be exercised end-to-end against a multi-site
   cluster. On a single node the `add` operation is rejected by the server.

## Example Usage

```terraform
resource "rustfs_site_replication" "peer" {
  name       = "peer"
  endpoint   = "http://peer.example.com:9001"
  access_key = "rustfsadmin"
  secret_key = "rustfsadmin"
}
```

## Schema

### Required

- `access_key` (String) Access key of the peer site
- `endpoint` (String) Endpoint of the peer site
- `name` (String) Unique name of the replication peer. Changing this forces recreation.
- `secret_key` (String, Sensitive) Secret key of the peer site

### Optional

- `ca_cert_pem` (String, Sensitive) Custom CA certificate (PEM) for the peer connection
- `skip_tls_verify` (Boolean) Skip TLS certificate verification when connecting to the peer. Defaults to false.

## Import

Import is supported using the peer name:

```
terraform import rustfs_site_replication.peer peer
```
