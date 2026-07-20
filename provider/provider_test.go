// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"rustfs": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
}

func testAccProviderConfig() string {
	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	if endpoint == "" {
		endpoint = "rustfs:9001"
	}
	user := os.Getenv("RUSTFS_USER")
	if user == "" {
		user = "rustfsadmin"
	}
	secret := os.Getenv("RUSTFS_SECRET")
	if secret == "" {
		secret = "rustfsadmin"
	}
	return `provider "rustfs" {
  endpoint      = "` + endpoint + `"
  access_key    = "` + user + `"
  access_secret = "` + secret + `"
  ssl           = false
}
`
}
