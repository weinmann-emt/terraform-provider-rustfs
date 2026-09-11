package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIAMPolicyDataSource(t *testing.T) {
	name := fmt.Sprintf("tf-test-policy-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_policy" "test" {
  name = "%s"
  statement = [{
    effect    = "Allow"
    action    = ["s3:GetObject"]
    resource = ["arn:aws:s3:::accbucket/*"]
  }]
}
data "rustfs_iam_policy" "test" {
  name = rustfs_policy.test.name
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.rustfs_iam_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.rustfs_iam_policy.test", "version", "2012-10-17"),
					resource.TestCheckResourceAttr("data.rustfs_iam_policy.test", "statement.#", "1"),
					resource.TestCheckResourceAttr("data.rustfs_iam_policy.test", "statement.0.effect", "Allow"),
				),
			},
		},
	})
}
