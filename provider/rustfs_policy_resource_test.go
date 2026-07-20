package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestAccPolicyResource_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-policy-%d", acctest.RandInt())
	resourceName := "rustfs_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func testAccPolicyConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_policy" "test" {
  name = "%s"
  statement = [{
    effect    = "Allow"
    action    = ["s3:GetObject", "s3:ListBucket"]
    ressource = ["arn:aws:s3:::*"]
  }]
}
`, name)
}

func testAccCheckPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		policyName := rs.Primary.Attributes["name"]
		if policyName == "" {
			return fmt.Errorf("no policy name set")
		}
		client := testAccPolicyClient()
		_, err := client.ReadPolicy(policyName)
		if err != nil {
			return fmt.Errorf("policy not found: %s", err)
		}
		return nil
	}
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	client := testAccPolicyClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_policy" {
			continue
		}
		policyName := rs.Primary.Attributes["name"]
		if policyName == "" {
			continue
		}
		_, err := client.ReadPolicy(policyName)
		if err == nil {
			return fmt.Errorf("policy %s still exists", policyName)
		}
	}
	return nil
}

func testAccPolicyClient() rustfs.RustfsAdmin {
	return rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})
}
