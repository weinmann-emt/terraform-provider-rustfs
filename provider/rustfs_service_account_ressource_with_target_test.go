// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	// "github.com/hashicorp/terraform-plugin-framework/provider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Due to TestAcc this is _only_ an acceptance test
func TestAccAserviceAccountWithTargetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
					resource "rustfs_serviceaccount" "target" {
						access_key = "testuser2kk"
						secret_key = "superSecret"
						description = "readonly"
						name = "createdaccount2"
						user = "testuser2"
						depends_on = [rustfs_user.target]
					}

					resource "rustfs_user" "target" {
						access_key = "testuser2"
						secret_key = "superSecret"
						policy = "readonly"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_serviceaccount.target", "access_key", "testuser2kk"),
					resource.TestCheckResourceAttr("rustfs_serviceaccount.target", "user", "testuser2"),
				)},
		}})
}
