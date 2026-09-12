---
page_title: "rustfs_ldap_policy_attachment Resource - rustfs"
description: |-
  Attach a canned policy to an LDAP user or group
---

# rustfs_ldap_policy_attachment (Resource)

Attach a canned policy to an LDAP user or group. The policy is detached when the resource is destroyed.

The target user or group must exist on the RustFS cluster, which requires an LDAP identity provider (IdP) to be configured and the LDAP directory to be synced. Set `is_group = true` to attach to an LDAP group instead of an LDAP user.

## Example Usage

```terraform
resource "rustfs_ldap_policy_attachment" "alice_readwrite" {
  user_or_group = "uid=alice,ou=people,dc=example,dc=com"
  policy        = "readwrite"
}

resource "rustfs_ldap_policy_attachment" "engineers_readonly" {
  user_or_group = "cn=engineers,ou=groups,dc=example,dc=com"
  policy        = "readonly"
  is_group      = true
}
```

## Schema

### Required

- `policy` (String) Name of the canned policy. Changing this forces a new attachment.
- `user_or_group` (String) LDAP user or group (distinguished name) the policy is attached to. Changing this forces a new attachment.

### Optional

- `is_group` (Boolean) Whether `user_or_group` is an LDAP group. Defaults to `false` (LDAP user).

### Read-Only

- `id` (String) Composite identifier in the form `user_or_group/policy`.

## Import

Import is supported using the composite `<user_or_group>/<policy>` identifier, optionally suffixed with the `is_group` flag:

```
terraform import rustfs_ldap_policy_attachment.test uid=alice,ou=people,dc=example,dc=com/readwrite
terraform import rustfs_ldap_policy_attachment.group cn=engineers,ou=groups,dc=example,dc=com/readonly/true
```

## Acceptance Testing

The acceptance tests require an LDAP identity provider to be configured on the RustFS cluster and the target LDAP user/group to exist. They skip gracefully when no LDAP IdP is configured.

Without an LDAP IdP you can still exercise the API end-to-end by pointing the tests at a regular (builtin) user that exists on the cluster:

```
# create the user, e.g. via the rustfs_user resource, then:
RUSTFS_LDAP_POLICY_USER_TARGET=my-test-user \
  TF_ACC=1 RUSTFS_ENDPOINT=localhost:9001 RUSTFS_USER=rustfsadmin RUSTFS_SECRET=rustfsadmin \
  go test -v -failfast -timeout 10m ./provider -run "TestAccLdapPolicyAttachment" -count=1
```

With a configured LDAP IdP, set the target environment variables to real distinguished names instead:

```
RUSTFS_LDAP_POLICY_USER_TARGET="uid=alice,ou=people,dc=example,dc=com" \
RUSTFS_LDAP_POLICY_GROUP_TARGET="cn=engineers,ou=groups,dc=example,dc=com" \
  TF_ACC=1 RUSTFS_ENDPOINT=localhost:9001 RUSTFS_USER=rustfsadmin RUSTFS_SECRET=rustfsadmin \
  go test -v -failfast -timeout 10m ./provider -run "TestAccLdapPolicyAttachment" -count=1
```