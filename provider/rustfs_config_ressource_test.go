package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestAccConfigResource(t *testing.T) {
	scope := fmt.Sprintf("notify_webhook:tfaccept%d", acctest.RandInt())
	resourceName := "rustfs_config.test"

	// Pre-cleanup in case a previous interrupted run left the target behind.
	client := rustfs.New(&rustfs.RustfsAdminConfig{
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		Ssl:          false,
	})
	_ = client.DeleteConfig(scope)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConfig(scope, "http://127.0.0.1:18080/rustfs/events"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "sub_system", scope),
					resource.TestCheckResourceAttr(resourceName, "settings.endpoint", "http://127.0.0.1:18080/rustfs/events"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccConfigConfig(scope, "http://127.0.0.1:18080/rustfs/events_v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "settings.endpoint", "http://127.0.0.1:18080/rustfs/events_v2"),
				),
			},
		},
	})
}

func testAccConfigConfig(scope, endpoint string) string {
	return fmt.Sprintf(testAccProviderConfig()+`
resource "rustfs_config" "test" {
  sub_system = "%s"
  settings = {
    endpoint = "%s"
  }
}
`, scope, endpoint)
}

// testAccCheckConfigDestroy verifies the config scope was removed so the
// server configuration is left unchanged when the test exits.
func testAccCheckConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_config" {
			continue
		}
		scope := rs.Primary.Attributes["sub_system"]
		client := rustfs.New(&rustfs.RustfsAdminConfig{
			AccessKey:    os.Getenv("RUSTFS_USER"),
			AccessSecret: os.Getenv("RUSTFS_SECRET"),
			Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
			Ssl:          false,
		})
		_, err := client.GetConfig(scope)
		if err == nil {
			return fmt.Errorf("config scope %s still exists after destroy", scope)
		}
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}
	return nil
}
