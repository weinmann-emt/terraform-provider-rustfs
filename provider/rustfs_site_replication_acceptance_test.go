package provider

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// TestAccSiteReplicationResource exercises the rustfs_site_replication resource
// against a running RustFS cluster. A single-node server cannot form a
// site-replication group (the add endpoint rejects loopback endpoints), so the
// test skips gracefully when the server rejects the add. Full end-to-end
// validation requires a multi-site cluster.
func TestAccSiteReplicationResource(t *testing.T) {
	testAccPreCheck(t)
	checkSiteReplicationAvailable(t)
	resourceName := "rustfs_site_replication.peer"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "rustfs_site_replication" "peer" {
  name       = "peer"
  endpoint   = "http://localhost:9002"
  access_key = "rustfsadmin"
  secret_key = "rustfsadmin"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "peer"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", "http://localhost:9002"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_key", "ca_cert_pem"},
			},
		},
	})
}

// checkSiteReplicationAvailable attempts a real add against the configured
// endpoint and skips the test when the server rejects it (single-node servers
// cannot form a site-replication group). This confirms the add endpoint is
// reachable while allowing the suite to run on a single instance.
func checkSiteReplicationAvailable(t *testing.T) {
	t.Helper()
	client := rustfs.New(&rustfs.RustfsAdminConfig{
		AccessKey:    rustfsadminAccessKey(),
		AccessSecret: rustfsadminSecretKey(),
		Endpoint:     rustfsadminEndpoint(),
		Ssl:          false,
	})
	err := client.SiteReplicationAdd(rustfs.SiteReplicationSite{
		Name:      "tf-acc-probe",
		Endpoint:  "http://localhost:9002",
		AccessKey: "rustfsadmin",
		SecretKey: "rustfsadmin",
	})
	if err == nil {
		_ = client.SiteReplicationRemove([]string{"tf-acc-probe"})
		return
	}
	if strings.Contains(err.Error(), "InvalidRequest") || strings.Contains(err.Error(), "loopback") {
		t.Skipf("site replication unavailable on single-node server: %v", err)
	}
	t.Fatalf("unexpected error probing site replication: %v", err)
}

func rustfsadminEndpoint() string {
	return envOr("RUSTFS_ENDPOINT", "")
}

func rustfsadminAccessKey() string {
	return envOr("RUSTFS_USER", "")
}

func rustfsadminSecretKey() string {
	return envOr("RUSTFS_SECRET", "")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
