package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
