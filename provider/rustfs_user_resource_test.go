package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestAccUserResource_basic(t *testing.T) {
	accessKey := fmt.Sprintf("tf-test-user-%d", acctest.RandInt())
	resourceName := "rustfs_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(accessKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, "enabled"),
					resource.TestCheckResourceAttr(resourceName, "access_key", accessKey),
					resource.TestCheckResourceAttr(resourceName, "status", "enabled"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_key"},
			},
		},
	})
}

func testAccUserConfig(accessKey string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_user" "test" {
  access_key = "%s"
  secret_key = "superSecret123!"
}
`, accessKey)
}

func testAccCheckUserExists(n string, expectedStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		accessKey := rs.Primary.Attributes["access_key"]
		if accessKey == "" {
			return fmt.Errorf("no access_key set")
		}

		client := testAccRustClient()
		user, err := client.ReadUserAccount(accessKey)
		if err != nil {
			return fmt.Errorf("user not found: %s", err)
		}
		if expectedStatus != "" && user.Status != expectedStatus {
			return fmt.Errorf("expected status=%s, got %s", expectedStatus, user.Status)
		}
		return nil
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	client := testAccRustClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_user" {
			continue
		}
		accessKey := rs.Primary.Attributes["access_key"]
		if accessKey == "" {
			continue
		}
		_, err := client.ReadUserAccount(accessKey)
		if err == nil {
			return fmt.Errorf("user %s still exists", accessKey)
		}
	}
	return nil
}

func testAccRustClient() rustfs.RustfsAdmin {
	config := &rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	}
	return rustfs.New(config)
}
