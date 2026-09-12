package provider

import (
	"context"
	"fmt"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBucketLifecycleConfigurationRessourceExpandedSchema(t *testing.T) {
	r := NewBucketLifecycleConfigurationRessource()
	req := frameworkresource.SchemaRequest{}
	resp := frameworkresource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	ruleBlock, ok := resp.Schema.Blocks["rule"].(schema.ListNestedBlock)
	if !ok {
		t.Fatal("rule block missing or not ListNestedBlock")
	}
	for _, want := range []string{"transition", "noncurrent_version_expiration", "noncurrent_version_transition", "abort_incomplete_multipart_upload", "expiration", "filter"} {
		if _, ok := ruleBlock.NestedObject.Blocks[want]; !ok {
			t.Errorf("missing rule sub-block %q", want)
		}
	}

	expBlock, ok := ruleBlock.NestedObject.Blocks["expiration"].(schema.SingleNestedBlock)
	if !ok {
		t.Fatal("expiration block missing or not SingleNestedBlock")
	}
	for _, want := range []string{"days", "date", "expired_object_delete_marker"} {
		if _, ok := expBlock.Attributes[want]; !ok {
			t.Errorf("missing expiration attribute %q", want)
		}
	}
}

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

func TestAccBucketLifecycleConfigurationResource_full(t *testing.T) {
	name := fmt.Sprintf("tf-test-lc-full-%d", acctest.RandInt())
	resourceName := "rustfs_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleFullConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule.0.id", "expire-days"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.filter.prefix", "logs/"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.expiration.days", "30"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.noncurrent_version_expiration.noncurrent_days", "90"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.abort_incomplete_multipart_upload.days_after_initiation", "7"),

					resource.TestCheckResourceAttr(resourceName, "rule.1.id", "expire-marker"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.filter.prefix", "tombstones/"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.expiration.expired_object_delete_marker", "true"),
				),
			},
		},
	})
}

func testAccLifecycleFullConfig(bucket string) string {
	return fmt.Sprintf(testAccProviderConfig()+`
resource "rustfs_bucket" "test" {
  name = "%s"
}

resource "rustfs_bucket_lifecycle_configuration" "test" {
  bucket = rustfs_bucket.test.name

  rule {
    id     = "expire-days"
    status = "Enabled"

    filter {
      prefix = "logs/"
    }

    expiration {
      days = 30
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }

  rule {
    id     = "expire-marker"
    status = "Enabled"

    filter {
      prefix = "tombstones/"
    }

    expiration {
      expired_object_delete_marker = true
    }
  }
}
`, bucket)
}
