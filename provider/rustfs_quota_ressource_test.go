package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQuotaResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "rustfs_bucket" "test2" {
  name = "somebucket2"
}
resource "rustfs_quota" "test" {
  bucket = "somebucket2"
  quota = 100000
  depends_on = [rustfs_bucket.test2]
}
`, Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_bucket.test2", "name", "somebucket2"),
					resource.TestCheckResourceAttr("rustfs_quota.test", "quota", "100000"),
				)},
		}})
}
