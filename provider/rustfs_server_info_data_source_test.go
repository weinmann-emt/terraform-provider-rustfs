package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerInfoDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "rustfs_server_info" "cluster" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rustfs_server_info.cluster", "mode"),
					resource.TestCheckResourceAttrSet("data.rustfs_server_info.cluster", "deployment_id"),
					resource.TestCheckResourceAttrSet("data.rustfs_server_info.cluster", "backend_type"),
					resource.TestCheckResourceAttrSet("data.rustfs_server_info.cluster", "servers.#"),
					resource.TestCheckResourceAttrSet("data.rustfs_server_info.cluster", "raw_json"),
				),
			},
		},
	})
}
