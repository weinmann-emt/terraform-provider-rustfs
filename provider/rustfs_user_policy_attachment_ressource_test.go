package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserPolicyAttachmentResource(t *testing.T) {
	userName := fmt.Sprintf("tf-test-upa-%d", acctest.RandInt())
	policyName := fmt.Sprintf("accpolicy-%d", acctest.RandInt())
	resourceName := "rustfs_user_policy_attachment.test"
	attachmentID := userName + "/" + policyName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentConfig(userName, policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user", userName),
					resource.TestCheckResourceAttr(resourceName, "policy", policyName),
					resource.TestCheckResourceAttr(resourceName, "id", attachmentID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     attachmentID,
			},
		},
	})
}

func testAccUserPolicyAttachmentConfig(userName, policyName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_policy" "test" {
  name = "%s"
  statement = [{
    effect    = "Allow"
    action    = ["s3:GetObject"]
    ressource = ["arn:aws:s3:::accbucket/*"]
  }]
}
resource "rustfs_user" "test" {
  access_key = "%s"
  secret_key = "superSecret123!"
  policy     = rustfs_policy.test.name
}
resource "rustfs_user_policy_attachment" "test" {
  user   = rustfs_user.test.access_key
  policy = rustfs_policy.test.name
}
`, policyName, userName)
}

func testAccCheckUserPolicyAttachmentDestroy(s *terraform.State) error {
	client := testAccRustClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_user_policy_attachment" {
			continue
		}
		user := rs.Primary.Attributes["user"]
		if user == "" {
			continue
		}
		if _, err := client.ReadUserAccount(user); err == nil {
			return fmt.Errorf("user %s still exists", user)
		}
	}
	return nil
}
