package provider

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

const (
	testAccLdapPolicyUserTargetDefault  = "uid=terraform-acc-test,ou=people,dc=example,dc=com"
	testAccLdapPolicyGroupTargetDefault = "cn=terraform-acc-test,ou=groups,dc=example,dc=com"
)

func testAccLdapPolicyUserTarget() string {
	if v := os.Getenv("RUSTFS_LDAP_POLICY_USER_TARGET"); v != "" {
		return v
	}
	return testAccLdapPolicyUserTargetDefault
}

func testAccLdapPolicyGroupTarget() string {
	if v := os.Getenv("RUSTFS_LDAP_POLICY_GROUP_TARGET"); v != "" {
		return v
	}
	return testAccLdapPolicyGroupTargetDefault
}

// testAccLdapPolicyPreCheck probes the server to confirm the LDAP policy
// endpoint is reachable and the target user_or_group exists. Without a
// configured LDAP identity provider (or without the target entity) the attach
// fails with a "not exist" error, in which case the test skips gracefully.
func testAccLdapPolicyPreCheck(t *testing.T, target string, isGroup bool) {
	testAccPreCheck(t)
	client := rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
		Ssl:          false,
	})
	req := rustfs.LDAPPolicyAttachment{
		UserOrGroup: target,
		PolicyName:  "readwrite",
		IsGroup:     isGroup,
	}
	if err := client.AttachLDAPPolicy(req); err != nil {
		if strings.Contains(err.Error(), "not exist") || strings.Contains(err.Error(), "does not exist") {
			t.Skipf("no LDAP identity provider configured (or target %q does not exist); "+
				"skipping LDAP policy attachment acceptance test. See docs/resources/ldap_policy_attachment.md "+
				"on configuring an LDAP IdP or pointing RUSTFS_LDAP_POLICY_USER_TARGET/"+
				"RUSTFS_LDAP_POLICY_GROUP_TARGET at an existing user/group.", target)
		}
		t.Fatalf("unexpected error probing LDAP policy attachment: %v", err)
	}
	if err := client.DetachLDAPPolicy(req); err != nil {
		t.Fatalf("failed to clean up LDAP policy probe: %v", err)
	}
}

func TestAccLdapPolicyAttachment_basic(t *testing.T) {
	target := testAccLdapPolicyUserTarget()
	resourceName := "rustfs_ldap_policy_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccLdapPolicyPreCheck(t, target, false) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLdapPolicyAttachmentConfig("test", target, "readwrite", "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_or_group", target),
					resource.TestCheckResourceAttr(resourceName, "policy", "readwrite"),
					resource.TestCheckResourceAttr(resourceName, "is_group", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     target + "/readwrite",
			},
		},
	})
}

func TestAccLdapPolicyAttachment_group(t *testing.T) {
	target := testAccLdapPolicyGroupTarget()
	resourceName := "rustfs_ldap_policy_attachment.group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccLdapPolicyPreCheck(t, target, true) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLdapPolicyAttachmentConfig("group", target, "readonly", "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_or_group", target),
					resource.TestCheckResourceAttr(resourceName, "policy", "readonly"),
					resource.TestCheckResourceAttr(resourceName, "is_group", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     target + "/readonly/true",
			},
		},
	})
}

func testAccLdapPolicyAttachmentConfig(resourceLabel, target, policy, isGroup string) string {
	return testAccProviderConfig() + `
resource "rustfs_ldap_policy_attachment" "` + resourceLabel + `" {
  user_or_group = "` + target + `"
  policy        = "` + policy + `"
  is_group      = ` + isGroup + `
}
`
}
