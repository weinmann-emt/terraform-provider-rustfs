package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "rustfs_policy" "test" {
  name = "providerpolicy"
	statement = [{
				effect = "Allow"
				action = ["s3:*"]
				ressource = ["arn:aws:s3:::*"]
	}]
}
`, Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_policy.test", "name", "providerpolicy"),
				)},
		}})
}
