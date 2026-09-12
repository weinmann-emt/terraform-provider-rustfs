package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHealthInfoDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "rustfs_health_info" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rustfs_health_info.test", "health_info"),
					resource.TestCheckResourceAttrSet("data.rustfs_health_info.test", "obd_info"),
					resource.TestCheckResourceAttrSet("data.rustfs_health_info.test", "version"),
					resource.TestCheckResourceAttrSet("data.rustfs_health_info.test", "region"),
					resource.TestCheckResourceAttrSet("data.rustfs_health_info.test", "timestamp"),
				),
			},
		},
	})
}
