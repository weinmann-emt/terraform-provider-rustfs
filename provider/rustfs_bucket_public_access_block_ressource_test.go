package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBucketPublicAccessBlockRessourceSchema(t *testing.T) {
	r := NewBucketPublicAccessBlockRessource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.TODO(), resource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.GetAttributes()
	for _, want := range []string{"bucket", "id", "block_public_acls", "ignore_public_acls", "block_public_policy", "restrict_public_buckets"} {
		if _, ok := attrs[want]; !ok {
			t.Errorf("missing attribute %q", want)
		}
	}
}

func TestBucketPublicAccessBlockRessourceMetadata(t *testing.T) {
	r := NewBucketPublicAccessBlockRessource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.TODO(), resource.MetadataRequest{ProviderTypeName: "rustfs"}, resp)

	if resp.TypeName != "rustfs_bucket_public_access_block" {
		t.Errorf("expected rustfs_bucket_public_access_block, got %s", resp.TypeName)
	}
}

func TestAccBucketPublicAccessBlock_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-pab-%d", acctest.RandInt())
	resourceName := "rustfs_bucket_public_access_block.test"

	tfresource.ParallelTest(t, tfresource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []tfresource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, true, false, true, false),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr(resourceName, "bucket", name),
					tfresource.TestCheckResourceAttr(resourceName, "id", name),
					tfresource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
					tfresource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
					tfresource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
					tfresource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, false, true, true, true),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
					tfresource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
					tfresource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
					tfresource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
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

func testAccBucketPublicAccessBlockConfig(bucket string, blockPublicAcls, ignorePublicAcls, blockPublicPolicy, restrictPublicBuckets bool) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_bucket" "test" {
  name = "%s"
}

resource "rustfs_bucket_public_access_block" "test" {
  bucket                = rustfs_bucket.test.name
  block_public_acls     = %t
  ignore_public_acls    = %t
  block_public_policy   = %t
  restrict_public_buckets = %t
}
`, bucket, blockPublicAcls, ignorePublicAcls, blockPublicPolicy, restrictPublicBuckets)
}
