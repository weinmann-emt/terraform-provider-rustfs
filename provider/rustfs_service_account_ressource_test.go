package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Due to TestAcc this is _only_ an acceptance test.
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

					resource "rustfs_serviceaccount" "test_with_policy" {
						access_key = "testuser-policy"
						secret_key = "superSecret"
						description = "readonly"
						name = "createdaccount-with-policy"
						policy = {
							statement = [{
								action = ["s3:*"]
								effect = "Allow"
								resource = ["arn:aws:s3:::*"]
							}]
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_serviceaccount.test", "access_key", "testuser"),
					resource.TestCheckResourceAttr("rustfs_serviceaccount.test_with_policy", "policy.statement.0.resource.0", "arn:aws:s3:::*"),
				)},
			{
				//Update test
				Config: providerConfig + `
					resource "rustfs_serviceaccount" "test" {
						access_key = "testuser"
						secret_key = "insecureOne"
						description = "readonly"
						name = "createdaccount"
						policy = {
							statement = [{
								action = ["s3:*"]
								effect = "Allow"
								resource = ["arn:aws:s3:::test"]
							}]
						}
					}

					resource "rustfs_serviceaccount" "test_with_policy" {
						access_key = "testuser-policy"
						secret_key = "superSecret"
						description = "readonly"
						name = "createdaccount-with-policy"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_serviceaccount.test", "access_key", "testuser"),
					resource.TestCheckResourceAttr("rustfs_serviceaccount.test", "policy.statement.0.resource.0", "arn:aws:s3:::test"),
				)},
		}})
}
