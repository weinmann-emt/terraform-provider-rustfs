package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestAccBucketResource_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	resourceName := "rustfs_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
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

func TestAccBucketResource_update(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	name2 := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	resourceName := "rustfs_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig(name),
				Check:  resource.TestCheckResourceAttr(resourceName, "name", name),
			},
			{
				Config: testAccBucketConfig(name2),
				Check:  resource.TestCheckResourceAttr(resourceName, "name", name2),
			},
		},
	})
}

func testAccBucketConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_bucket" "test" {
  name = "%s"
}
`, name)
}

func testAccCheckBucketExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set")
		}

		client, err := testAccMinioClient()
		if err != nil {
			return err
		}

		exists, err := client.BucketExists(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error checking bucket: %s", err)
		}
		if !exists {
			return fmt.Errorf("bucket %s does not exist", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckBucketDestroy(s *terraform.State) error {
	client, err := testAccMinioClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_bucket" {
			continue
		}

		exists, err := client.BucketExists(context.Background(), rs.Primary.ID)
		if err != nil {
			minioErr, ok := err.(minio.ErrorResponse)
			if ok && (strings.Contains(minioErr.Code, "NotFound") || minioErr.StatusCode == 404) {
				continue
			}
			return fmt.Errorf("error checking bucket destruction: %s", err)
		}
		if exists {
			return fmt.Errorf("bucket %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccMinioClient() (*minio.Client, error) {
	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	accessKey := os.Getenv("RUSTFS_USER")
	secretKey := os.Getenv("RUSTFS_SECRET")
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
}
