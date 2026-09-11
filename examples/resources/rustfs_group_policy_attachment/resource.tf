resource "rustfs_group_policy_attachment" "developers_readwrite" {
  group  = "developers"
  policy = "readwrite"
}