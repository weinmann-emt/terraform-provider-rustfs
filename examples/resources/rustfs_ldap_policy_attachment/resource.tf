resource "rustfs_ldap_policy_attachment" "alice_readwrite" {
  user_or_group = "uid=alice,ou=people,dc=example,dc=com"
  policy        = "readwrite"
}