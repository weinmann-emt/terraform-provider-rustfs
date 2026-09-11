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

// TestAccLdapServiceAccount creates a service account scoped to an LDAP user.
// The endpoint only accepts target users that already exist in the identity
// store, which requires an LDAP IdP configured on the RustFS cluster (LDAP
// users are provisioned on login). When no such user exists the test skips
// gracefully; set RUSTFS_LDAP_TEST_USER_DN to point at a real LDAP user DN.
func TestAccLdapServiceAccount(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	accessKey := fmt.Sprintf("tf-ldap-sa-%d", acctest.RandInt())
	resourceName := "rustfs_ldap_service_account.test"
	dn := envOrDefault("RUSTFS_LDAP_TEST_USER_DN", "uid=alice,ou=people,dc=example,dc=com")

	client := rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})

	// Probe the endpoint: if the LDAP user is not provisioned, the server
	// rejects the create with "target user not exist". Skip in that case so
	// the acceptance run stays green on clusters without an LDAP IdP.
	probeKey := accessKey + "-probe"
	err := client.CreateLDAPServiceAccount(rustfs.ServiceAccount{
		AccessKey:  probeKey,
		SecretKey:  "probeSecret",
		Name:       "ldap-probe",
		TargetUser: dn,
	})
	if err != nil {
		t.Skipf("skipping: no LDAP IdP configured or LDAP user %q missing on the cluster: %v", dn, err)
	}
	if delErr := client.DeleteServiceAccount(rustfs.ServiceAccount{AccessKey: probeKey}); delErr != nil {
		t.Fatalf("could not clean up probe service account: %v", delErr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckLdapServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLdapServiceAccountConfig(accessKey, dn, "ldap-token"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_key", accessKey),
					resource.TestCheckResourceAttr(resourceName, "user", dn),
					resource.TestCheckResourceAttr(resourceName, "name", "ldap-token"),
				),
			},
			{
				Config: testAccLdapServiceAccountConfig(accessKey, dn, "renamed-token"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "renamed-token"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        accessKey,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "access_key",
				ImportStateVerifyIgnore:              []string{"secret_key"},
			},
		},
	})
}

func testAccLdapServiceAccountConfig(accessKey, user, name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_ldap_service_account" "test" {
  access_key = "%s"
  secret_key = "superSecret123!"
  name       = "%s"
  user       = %q
}
`, accessKey, name, user)
}

func testAccCheckLdapServiceAccountDestroy(s *terraform.State) error {
	client := rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_ldap_service_account" {
			continue
		}
		accessKey := rs.Primary.Attributes["access_key"]
		if accessKey == "" {
			continue
		}
		_, err := client.ReadServiceAccount(accessKey)
		if err == nil {
			return fmt.Errorf("service account %s still exists", accessKey)
		}
	}
	return nil
}
