package provider

import (
	"testing"

	// "github.com/hashicorp/terraform-plugin-framework/provider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Due to TestAcc this is _only_ an acceptance test
func TestAccAserviceAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
					resource "rustfs_serviceaccount" "test" {
						access_key = "testuser"
						secret_key = "superSecret"
						description = "readonly"
						name = "createdaccount"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_serviceaccount.test", "access_key", "testuser"),
				)},
			{
				//Update test
				Config: providerConfig + `
					resource "rustfs_serviceaccount" "test" {
						access_key = "testuser"
						secret_key = "insecureOne"
						description = "readonly"
						name = "createdaccount"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_serviceaccount.test", "access_key", "testuser"),
				)},
		}})
}
