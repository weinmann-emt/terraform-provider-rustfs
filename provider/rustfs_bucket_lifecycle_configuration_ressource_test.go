package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBucketLifecycleConfigurationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig + `
resource "rustfs_bucket" "test_bucket" {
  name = "test-lifecycle-bucket"
}

resource "rustfs_bucket_lifecycle_configuration" "test" {
  bucket = rustfs_bucket.test_bucket.name

  rule {
    id     = "rule1"
    status = "Enabled"

    filter {
      prefix = "logs/"
    }

    expiration {
      days = 30
    }
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "bucket", "test-lifecycle-bucket"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "id", "test-lifecycle-bucket"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.id", "rule1"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.status", "Enabled"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.filter.prefix", "logs/"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.expiration.days", "30"),
				),
			},
			// Update and Read
			{
				Config: providerConfig + `
resource "rustfs_bucket" "test_bucket" {
  name = "test-lifecycle-bucket"
}

resource "rustfs_bucket_lifecycle_configuration" "test" {
  bucket = rustfs_bucket.test_bucket.name

  rule {
    id     = "rule1"
    status = "Disabled"

    filter {
      prefix = "archive/"
    }

    expiration {
      days = 90
    }
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "bucket", "test-lifecycle-bucket"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.id", "rule1"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.status", "Disabled"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.filter.prefix", "archive/"),
					resource.TestCheckResourceAttr("rustfs_bucket_lifecycle_configuration.test", "rule.0.expiration.days", "90"),
				),
			},
		},
	})
}
