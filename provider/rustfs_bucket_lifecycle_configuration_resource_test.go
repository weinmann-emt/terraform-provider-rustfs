package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBucketLifecycleConfigurationResource_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-lc-%d", acctest.RandInt())
	resourceName := "rustfs_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleConfig(name, "Enabled", "logs/", 30),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule.0.id", "rule1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.prefix", "logs/"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.days", "30"),
				),
			},
			{
				Config: testAccLifecycleConfig(name, "Disabled", "archive/", 90),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule.0.status", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.prefix", "archive/"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.days", "90"),
				),
			},
		},
	})
}

func testAccLifecycleConfig(bucket, status, prefix string, days int) string {
	return fmt.Sprintf(testAccProviderConfig()+`
resource "rustfs_bucket" "test" {
  name = "%s"
}

resource "rustfs_bucket_lifecycle_configuration" "test" {
  bucket = rustfs_bucket.test.name

  rule {
    id     = "rule1"
    status = "%s"

    filter {
      prefix = "%s"
    }

    expiration {
      days = %d
    }
  }
}
`, bucket, status, prefix, days)
}
