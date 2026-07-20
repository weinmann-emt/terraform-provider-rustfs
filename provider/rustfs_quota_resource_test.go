package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQuotaResource_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-quota-%d", acctest.RandInt())
	resourceName := "rustfs_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQuotaConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "quota", "100000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccQuotaResource_update(t *testing.T) {
	name := fmt.Sprintf("tf-test-quota-%d", acctest.RandInt())
	resourceName := "rustfs_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQuotaConfigValue(name, 100000),
				Check:  resource.TestCheckResourceAttr(resourceName, "quota", "100000"),
			},
			{
				Config: testAccQuotaConfigValue(name, 200000),
				Check:  resource.TestCheckResourceAttr(resourceName, "quota", "200000"),
			},
		},
	})
}

func testAccQuotaConfig(name string) string {
	return testAccQuotaConfigValue(name, 100000)
}

func testAccQuotaConfigValue(name string, quota int) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_bucket" "test" {
  name = "%s"
}

resource "rustfs_quota" "test" {
  bucket     = rustfs_bucket.test.name
  quota      = %d
  depends_on = [rustfs_bucket.test]
}
`, name, quota)
}
