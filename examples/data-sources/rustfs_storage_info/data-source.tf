terraform {
  required_providers {
    rustfs = {
      source = "weinmann/rustfs"
    }
  }
}

provider "rustfs" {
  endpoint      = "rustfs:9001"
  access_key    = "rustfsadmin"
  access_secret = "rustfsadmin"
  ssl           = false
}

data "rustfs_storage_info" "cluster" {}

output "backend_type" {
  value = data.rustfs_storage_info.cluster.backend.backend_type
}

output "raw" {
  value = data.rustfs_storage_info.cluster.raw_json
}
