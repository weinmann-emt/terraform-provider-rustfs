package provider

import (
	"context"
	"fmt"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBucketTagsRessourceSchema(t *testing.T) {
	r := NewBucketTagsRessource()
	req := fwresource.SchemaRequest{}
	resp := fwresource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	for _, want := range []string{"bucket", "id", "tags"} {
		if _, ok := resp.Schema.Attributes[want]; !ok {
			t.Errorf("missing attribute %q", want)
		}
	}
}

func TestAccBucketTags_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-tags-%d", acctest.RandInt())
	resourceName := "rustfs_bucket_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketTagsConfig(name, `environment = "production"`, `team        = "platform"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "bucket", name),
					resource.TestCheckResourceAttr(resourceName, "id", name),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.team", "platform"),
				),
			},
			{
				Config: testAccBucketTagsConfig(name, `environment = "staging"`, `costcenter  = "cc-42"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "staging"),
					resource.TestCheckResourceAttr(resourceName, "tags.costcenter", "cc-42"),
					resource.TestCheckNoResourceAttr(resourceName, "tags.team"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     name,
			},
		},
	})
}

func testAccBucketTagsConfig(bucket, tag1, tag2 string) string {
	return fmt.Sprintf(testAccProviderConfig()+`
resource "rustfs_bucket" "test_bucket" {
  name = "%s"
}

resource "rustfs_bucket_tags" "test" {
  bucket = rustfs_bucket.test_bucket.name

  tags = {
    %s
    %s
  }
}
`, bucket, tag1, tag2)
}
