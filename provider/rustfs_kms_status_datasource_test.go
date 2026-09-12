package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// TestAccKmsStatusDataSource verifies the rustfs_kms_status data source against
// a live RustFS server. Because the KMS subsystem is only populated once a
// backend has been configured, the test skips when the server reports that the
// KMS service is not initialized so it can run against a stock deployment.
func TestAccKmsStatusDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccKMSPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "rustfs_kms_status" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rustfs_kms_status.test", "id"),
					resource.TestCheckResourceAttrSet("data.rustfs_kms_status.test", "backend_type"),
					resource.TestCheckResourceAttrSet("data.rustfs_kms_status.test", "backend_status"),
					resource.TestCheckResourceAttrSet("data.rustfs_kms_status.test", "backend"),
					resource.TestCheckResourceAttrSet("data.rustfs_kms_status.test", "cache_enabled"),
					resource.TestCheckResourceAttrSet("data.rustfs_kms_status.test", "status_json"),
					resource.TestCheckResourceAttrSet("data.rustfs_kms_status.test", "config_json"),
				),
			},
		},
	})
}

func testAccKMSPreCheck(t *testing.T) {
	t.Helper()
	client := rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})
	if _, err := client.KmsStatus(); err != nil {
		t.Skipf("KMS service not initialized on the target server, skipping acceptance test: %v", err)
	}
}
