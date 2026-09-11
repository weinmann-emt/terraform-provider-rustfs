---
page_title: "rustfs_audit_target Resource - rustfs"
description: |-
  Manage RustFS audit-log webhook/HTTP targets
---

# rustfs_audit_target (Resource)

Manage RustFS audit-log webhook/HTTP targets.

The audit target is addressed by its `target_type` (the audit subsystem, e.g.
`audit_webhook`) and `target_name`. Destroying this resource resets the target
to its default configuration via the admin `reset` endpoint (the admin API has
no plain delete for targets).

The admin list endpoint returns only identity and runtime health metadata for
each target; it does not echo the configured values (endpoint, auth token,
comment, ...). Config fields are therefore preserved by the provider and are
never refreshed from the server on read. `auth_token` is sensitive and never
logged.

## Example Usage

```terraform
resource "rustfs_audit_target" "test" {
  target_type = "audit_webhook"
  target_name = "terraform"
  endpoint    = "https://hooks.example.com/webhook/terraform"
  auth_token  = "superSecret"
  comment     = "managed by terraform"
}
```

## Schema

### Required

- `target_name` (String) Name of the audit target. Changing this forces recreation.

### Optional

- `target_type` (String) Audit target subsystem. Defaults to `audit_webhook`. Valid values: `audit_webhook`, `audit_kafka`, `audit_amqp`, `audit_mqtt`, `audit_mysql`, `audit_nats`, `audit_postgres`, `audit_pulsar`, `audit_redis`.
- `endpoint` (String) Endpoint URL the audit target posts to (webhook).
- `auth_token` (String, Sensitive) Optional bearer token for the audit target (webhook). Not refreshed from the server on read.
- `comment` (String) Optional comment describing the audit target.
- `queue_limit` (Number) Optional in-memory queue limit for the audit target.
- `queue_dir` (String) Optional absolute path of the on-disk queue directory for the audit target.
- `client_cert` (String, Sensitive) Optional client certificate for mTLS to the audit target.
- `client_key` (String, Sensitive) Optional client private key for mTLS to the audit target.
- `client_ca` (String, Sensitive) Optional CA certificate used to verify the audit target.
- `skip_tls_verify` (Boolean) Skip TLS certificate verification for the audit target. Defaults to `false`.

### Read-Only

- `health_state` (String) Runtime health state of the audit target (offline/online/...).
- `health_reason` (String) Runtime health reason for the audit target.
- `status` (String) Runtime status of the audit target.

## Notes

The RustFS audit module must be enabled on the server (set
`RUSTFS_AUDIT_ENABLE=true` or enable it from the console) before audit targets
can be managed; otherwise the admin API rejects target operations. Webhook
endpoints must also be allowed by the server's outbound policy
(`RUSTFS_OUTBOUND_ALLOW_ORIGINS`) — loopback/private addresses are rejected by
default.

## Import

Import is supported using the `<target_type>/<target_name>` pair:

```
terraform import rustfs_audit_target.test audit_webhook/terraform
```

Because the admin list endpoint does not echo target configuration values, only
`target_type` and `target_name` are populated on import; config attributes
(`endpoint`, `auth_token`, `comment`, ...) must be set in configuration after
import.