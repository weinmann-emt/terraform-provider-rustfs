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

func TestAccBucketDurability_basic(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-durability-%d", acctest.RandInt())
	resourceName := "rustfs_bucket_durability.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDurabilityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDurabilityConfig(bucketName, "strict"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketDurabilityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "mode", "strict"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        bucketName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "bucket",
			},
		},
	})
}

func TestAccBucketDurability_update(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-durability-%d", acctest.RandInt())
	resourceName := "rustfs_bucket_durability.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDurabilityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDurabilityConfig(bucketName, "strict"),
				Check:  resource.TestCheckResourceAttr(resourceName, "mode", "strict"),
			},
			{
				Config: testAccBucketDurabilityConfig(bucketName, "none"),
				Check:  resource.TestCheckResourceAttr(resourceName, "mode", "none"),
			},
		},
	})
}

func testAccBucketDurabilityConfig(bucket, mode string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_bucket" "test" {
  name = "%s"
}

resource "rustfs_bucket_durability" "test" {
  bucket     = rustfs_bucket.test.name
  mode       = "%s"
  depends_on = [rustfs_bucket.test]
}
`, bucket, mode)
}

func testAccCheckBucketDurabilityExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		bucketName := rs.Primary.Attributes["bucket"]
		if bucketName == "" {
			return fmt.Errorf("no bucket set")
		}
		client := testAccDurabilityClient()
		d, err := client.GetBucketDurability(bucketName)
		if err != nil {
			return fmt.Errorf("error reading bucket durability: %s", err)
		}
		if d.Mode != rs.Primary.Attributes["mode"] {
			return fmt.Errorf("durability mode %q does not match state %q", d.Mode, rs.Primary.Attributes["mode"])
		}
		return nil
	}
}

func testAccCheckBucketDurabilityDestroy(s *terraform.State) error {
	client := testAccDurabilityClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_bucket_durability" {
			continue
		}
		bucketName := rs.Primary.Attributes["bucket"]
		if bucketName == "" {
			continue
		}
		// The override is cleared on destroy; an empty mode means the bucket
		// inherits the process-wide durability mode again. A read error means
		// the bucket itself is already gone, which also counts as destroyed.
		d, err := client.GetBucketDurability(bucketName)
		if err != nil {
			continue
		}
		if d.Mode != "" {
			return fmt.Errorf("durability override for bucket %s still exists", bucketName)
		}
	}
	return nil
}

func testAccDurabilityClient() rustfs.RustfsAdmin {
	return rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})
}
