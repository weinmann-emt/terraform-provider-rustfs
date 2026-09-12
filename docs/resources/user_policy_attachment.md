---
page_title: "rustfs_user_policy_attachment Resource - rustfs"
description: |-
  Attach a canned IAM policy to a user
---

# rustfs_user_policy_attachment (Resource)

Attach a canned IAM policy to a user. The policy is detached when the resource is destroyed.

RustFS stores the policies of a user as a list, so a user can hold more than one policy. This
resource appends a single policy to that list on create and removes it on destroy, leaving any
other attached policies untouched.

## Interaction with rustfs_user.policy

`rustfs_user.policy` manages the user's primary policy. It is applied through the
`set-user-or-group-policy` endpoint, which **replaces** the whole policy mapping, so a user only
holds the single primary policy via that attribute.

`rustfs_user_policy_attachment` adds additional policies on top by calling the builtin-policy
attach/detach endpoints (`POST idp/builtin/policy/attach` and `POST idp/builtin/policy/detach`),
which append/remove a single policy without touching the rest.

The two resources are independent: removing an attachment never touches `rustfs_user.policy` and
vice versa. Changing `rustfs_user.policy` replaces the user, which also removes all its
attachments. Recommended usage: keep the first policy in `rustfs_user.policy` and express every
additional policy as an attachment (or model all policies as attachments).

## Example Usage

```terraform
resource "rustfs_user_policy_attachment" "alice_readwrite" {
  user   = rustfs_user.alice.access_key
  policy = "readwrite"
}
```

## Schema

### Required

- `policy` (String) Name of the canned policy. Changing this forces a new attachment.
- `user` (String) Access key of the IAM user. Changing this forces a new attachment.

### Read-Only

- `id` (String) Composite identifier in the form `user/policy`.

## Import

Import is supported using the composite `<user>/<policy>` identifier:

```
terraform import rustfs_user_policy_attachment.test alice/readwrite
```