package provider

import (
	"fmt"
	"slices"
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

// Destruction is verified against the group's policy mapping: after the
// attachment is removed the group must either be gone or no longer carry the
// policy.
func testAccCheckGroupPolicyAttachmentDestroy(s *terraform.State) error {
	client := testAccRustClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_group_policy_attachment" {
			continue
		}
		group := rs.Primary.Attributes["group"]
		policy := rs.Primary.Attributes["policy"]
		if group == "" {
			continue
		}
		policies, err := client.GetGroupPolicies(group)
		if err != nil {
			continue
		}
		if slices.Contains(policies, policy) {
			return fmt.Errorf("group %s still has policy %q attached", group, policy)
		}
	}
	return nil
}

func TestAccGroupPolicyAttachmentResource_multiple(t *testing.T) {
	groupName := fmt.Sprintf("tf-test-group-%d", acctest.RandInt())
	policyA := fmt.Sprintf("tf-test-policy-a-%d", acctest.RandInt())
	policyB := fmt.Sprintf("tf-test-policy-b-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentMultipleConfig(groupName, policyA, policyB),
				Check:  testAccCheckGroupHasPolicies(groupName, policyA, policyB),
			},
			{
				// Detach one policy outside Terraform; the next plan must
				// detect the drift and re-attach it.
				PreConfig: func() {
					client := testAccRustClient()
					if err := client.DetachGroupPolicy(groupName, policyA); err != nil {
						t.Fatalf("detaching out of band: %v", err)
					}
				},
				Config:             testAccGroupPolicyAttachmentMultipleConfig(groupName, policyA, policyB),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccGroupPolicyAttachmentMultipleConfig(groupName, policyA, policyB),
				Check:  testAccCheckGroupHasPolicies(groupName, policyA, policyB),
			},
		},
	})
}

func testAccCheckGroupHasPolicies(group string, want ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccRustClient()
		policies, err := client.GetGroupPolicies(group)
		if err != nil {
			return err
		}
		for _, p := range want {
			if !slices.Contains(policies, p) {
				return fmt.Errorf("group %s: expected policy %q attached, got %v", group, p, policies)
			}
		}
		return nil
	}
}

func testAccGroupPolicyAttachmentMultipleConfig(groupName, policyA, policyB string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_group" "test" {
  name = "%[1]s"
}

resource "rustfs_policy" "a" {
  name = "%[2]s"
  statement = [{
    effect    = "Allow"
    action    = ["s3:GetObject"]
    ressource = ["arn:aws:s3:::%[1]s/*"]
  }]
}

resource "rustfs_policy" "b" {
  name = "%[3]s"
  statement = [{
    effect    = "Allow"
    action    = ["s3:PutObject"]
    ressource = ["arn:aws:s3:::%[1]s/*"]
  }]
}

resource "rustfs_group_policy_attachment" "a" {
  group  = rustfs_group.test.name
  policy = rustfs_policy.a.name
}

resource "rustfs_group_policy_attachment" "b" {
  group  = rustfs_group.test.name
  policy = rustfs_policy.b.name
}
`, groupName, policyA, policyB)
}
