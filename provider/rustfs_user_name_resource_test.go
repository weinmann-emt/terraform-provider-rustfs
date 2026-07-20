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

func TestAccUserResource_nameDefault(t *testing.T) {
	accessKey := fmt.Sprintf("tf-test-name-%d", acctest.RandInt())
	resourceName := "rustfs_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserNameConfigExplicit(accessKey, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_key", accessKey),
					resource.TestCheckResourceAttr(resourceName, "name", accessKey),
				),
			},
		},
	})
}

func TestAccUserResource_nameExplicit(t *testing.T) {
	accessKey := fmt.Sprintf("tf-test-name-%d", acctest.RandInt())
	resourceName := "rustfs_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserNameConfigExplicit(accessKey, "My Beautiful User"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_key", accessKey),
					resource.TestCheckResourceAttr(resourceName, "name", "My Beautiful User"),
				),
			},
		},
	})
}

func TestAccUserResource_nameImport(t *testing.T) {
	accessKey := fmt.Sprintf("tf-test-name-%d", acctest.RandInt())
	resourceName := "rustfs_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserNameConfigExplicit(accessKey, ""),
				Check:  resource.TestCheckResourceAttr(resourceName, "name", accessKey),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        accessKey,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "access_key",
				ImportStateVerifyIgnore:              []string{"secret_key"},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", accessKey),
				),
			},
		},
	})
}

func testAccUserNameConfigExplicit(accessKey, name string) string {
	nameAttr := ""
	if name != "" {
		nameAttr = fmt.Sprintf(`name = "%s"`, name)
	}
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_user" "test" {
  access_key = "%s"
  secret_key = "superSecret123!"
  status     = "enabled"
  policy     = ""
  %s
}
`, accessKey, nameAttr)
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
	return rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})
}
