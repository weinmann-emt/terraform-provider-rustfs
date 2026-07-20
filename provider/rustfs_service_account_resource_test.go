package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccServiceAccountResource_basic(t *testing.T) {
	accessKey := fmt.Sprintf("tf-test-sa-%d", acctest.RandInt())
	resourceName := "rustfs_serviceaccount.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountConfig(accessKey, "test-service"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_key", accessKey),
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

func TestAccServiceAccountResource_update(t *testing.T) {
	accessKey := fmt.Sprintf("tf-test-sa-%d", acctest.RandInt())
	resourceName := "rustfs_serviceaccount.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountConfig(accessKey, "original-name"),
				Check:  resource.TestCheckResourceAttr(resourceName, "name", "original-name"),
			},
			{
				Config: testAccServiceAccountConfig(accessKey, "updated-name"),
				Check:  resource.TestCheckResourceAttr(resourceName, "name", "updated-name"),
			},
		},
	})
}

func testAccServiceAccountConfig(accessKey, name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_serviceaccount" "test" {
  access_key  = "%s"
  secret_key  = "superSecret123!"
  name        = "%s"
  description = "acceptance test service account"
}
`, accessKey, name)
}

func testAccCheckServiceAccountExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		client := testAccRustClient()
		_, err := client.ReadServiceAccount(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("service account not found: %s", err)
		}
		return nil
	}
}

func testAccCheckServiceAccountDestroy(s *terraform.State) error {
	client := testAccRustClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_serviceaccount" {
			continue
		}
		_, err := client.ReadServiceAccount(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("service account %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
