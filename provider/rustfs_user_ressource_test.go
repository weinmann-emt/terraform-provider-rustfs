package provider_test

import (
	"testing"

	// "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/weinmann-emt/terraform-provider-rustfs/provider"
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

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"rustfs": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
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
