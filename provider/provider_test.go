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

var requiredEnvVars = []string{
	"RUSTFS_ENDPOINT",
	"RUSTFS_USER",
	"RUSTFS_SECRET",
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	for _, env := range requiredEnvVars {
		if os.Getenv(env) == "" {
			t.Fatalf("%s must be set for acceptance tests", env)
		}
	}
}

func testAccProviderConfig() string {
	return `provider "rustfs" {
  endpoint      = "` + os.Getenv("RUSTFS_ENDPOINT") + `"
  access_key    = "` + os.Getenv("RUSTFS_USER") + `"
  access_secret = "` + os.Getenv("RUSTFS_SECRET") + `"
  ssl           = false
}
`
}
