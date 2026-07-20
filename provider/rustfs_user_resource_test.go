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
				Config: testAccUserConfig(accessKey, "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName),
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

func TestAccUserResource_status(t *testing.T) {
	accessKey := fmt.Sprintf("tf-test-user-%d", acctest.RandInt())
	resourceName := "rustfs_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(accessKey, "enabled"),
				Check:  resource.TestCheckResourceAttr(resourceName, "status", "enabled"),
			},
			{
				Config: testAccUserConfig(accessKey, "disabled"),
				Check:  resource.TestCheckResourceAttr(resourceName, "status", "disabled"),
			},
		},
	})
}

func testAccUserConfig(accessKey, status string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_user" "test" {
  access_key = "%s"
  secret_key = "superSecret123!"
  status     = "%s"
}
`, accessKey, status)
}

func testAccCheckUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		client := testAccRustClient()
		_, err := client.ReadUserAccount(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("user not found: %s", err)
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
		_, err := client.ReadUserAccount(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("user %s still exists", rs.Primary.ID)
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
