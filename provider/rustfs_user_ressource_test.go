package provider

import (

	// "github.com/hashicorp/terraform-plugin-framework/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the HashiCups client is properly configured.
	// It is also possible to use the HASHICUPS_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `terraform {
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
`
)

// Due to TestAcc this is _only_ an acceptance test
func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "rustfs_user" "test" {
  access_key = "testuser"
  secret_key = "superSecret"
	policy = "readonly"
}
`, Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_user.test", "access_key", "testuser"),
				)},
		}})
}
