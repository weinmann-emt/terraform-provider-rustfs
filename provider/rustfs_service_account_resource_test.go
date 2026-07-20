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
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        accessKey,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "access_key",
				ImportStateVerifyIgnore:              []string{"secret_key"},
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
		accessKey := rs.Primary.Attributes["access_key"]
		if accessKey == "" {
			return fmt.Errorf("no access_key set")
		}
		client := testAccSAClient()
		_, err := client.ReadServiceAccount(accessKey)
		if err != nil {
			return fmt.Errorf("service account not found: %s", err)
		}
		return nil
	}
}

func testAccCheckServiceAccountDestroy(s *terraform.State) error {
	client := testAccSAClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_serviceaccount" {
			continue
		}
		accessKey := rs.Primary.Attributes["access_key"]
		if accessKey == "" {
			continue
		}
		_, err := client.ReadServiceAccount(accessKey)
		if err == nil {
			return fmt.Errorf("service account %s still exists", accessKey)
		}
	}
	return nil
}

func testAccSAClient() rustfs.RustfsAdmin {
	return rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})
}
