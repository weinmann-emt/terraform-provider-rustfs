package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestQuotaDataSourceSchema(t *testing.T) {
	d := NewQuotaDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.TODO(), datasource.SchemaRequest{}, resp)

	if diags := resp.Diagnostics; diags.HasError() {
		t.Fatalf("schema diagnostics: %v", diags)
	}

	attrs := resp.Schema.GetAttributes()
	for _, attr := range []string{"bucket", "quota", "quota_type"} {
		if _, ok := attrs[attr]; !ok {
			t.Errorf("expected %s attribute", attr)
		}
	}
}

func TestQuotaDataSourceMetadata(t *testing.T) {
	d := NewQuotaDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.TODO(), datasource.MetadataRequest{ProviderTypeName: "rustfs"}, resp)

	if resp.TypeName != "rustfs_quota" {
		t.Errorf("expected rustfs_quota, got %s", resp.TypeName)
	}
}

func TestAccQuotaDataSource(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-quota-ds-%d", acctest.RandInt())
	resourceName := "rustfs_quota.test"
	dataSourceName := "data.rustfs_quota.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckQuotaAndBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQuotaDataSourceConfig(bucketName, 250000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket", resourceName, "bucket"),
					resource.TestCheckResourceAttr(dataSourceName, "quota", "250000"),
					resource.TestCheckResourceAttr(dataSourceName, "quota_type", "HARD"),
				),
			},
		},
	})
}

func testAccQuotaDataSourceConfig(bucket string, quota int) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_bucket" "test" {
  name = "%s"
}

resource "rustfs_quota" "test" {
  bucket     = rustfs_bucket.test.name
  quota      = %d
  depends_on = [rustfs_bucket.test]
}

data "rustfs_quota" "test" {
  bucket = rustfs_quota.test.bucket
}
`, bucket, quota)
}
