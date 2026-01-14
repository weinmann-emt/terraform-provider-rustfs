terraform {
	  required_providers {
    rustfs = {
      source = "weinmann/rustfs"
    }
  }
}
provider "rustfs" {
  endpoint = "rustfs:9001"
  access_key = "rustfsadmin"
  access_secret = "rustfsadmin"
  ssl= false
}

resource "rustfs_user" "test" {
  access_key = "testuser"
  secret_key = "superSecret"
}