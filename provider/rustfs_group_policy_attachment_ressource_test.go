package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccGroupPolicyAttachmentResource_basic(t *testing.T) {
	groupName := fmt.Sprintf("tf-test-group-%d", acctest.RandInt())
	policyName := fmt.Sprintf("tf-test-policy-%d", acctest.RandInt())
	resourceName := "rustfs_group_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentConfig(groupName, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "group", groupName),
					resource.TestCheckResourceAttr(resourceName, "policy", policyName),
					resource.TestCheckResourceAttr(resourceName, "id", groupName+"/"+policyName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     groupName + "/" + policyName,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGroupPolicyAttachmentConfig(groupName, policyName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_group" "test" {
  name = "%s"
}

resource "rustfs_policy" "test" {
  name = "%s"
  statement = [{
    effect    = "Allow"
    action    = ["s3:GetObject"]
    ressource = ["arn:aws:s3:::%s/*"]
  }]
}

resource "rustfs_group_policy_attachment" "test" {
  group  = rustfs_group.test.name
  policy = rustfs_policy.test.name
}
`, groupName, policyName, groupName)
}

// The admin API exposes no dedicated read-back for group policy attachments, so
// destruction is verified against the group's own policy mapping: after the
// attachment is removed the group must either be gone or carry no policy.
func testAccCheckGroupPolicyAttachmentDestroy(s *terraform.State) error {
	client := testAccRustClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_group_policy_attachment" {
			continue
		}
		group := rs.Primary.Attributes["group"]
		if group == "" {
			continue
		}
		info, err := client.GetGroup(group)
		if err != nil {
			continue
		}
		if info.Policy != "" {
			return fmt.Errorf("group %s still has policy %q attached", group, info.Policy)
		}
	}
	return nil
}
