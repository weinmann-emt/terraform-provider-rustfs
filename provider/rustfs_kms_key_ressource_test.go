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

const kmsKeyTestResourceName = "rustfs_kms_key.test"

func TestAccKmsKeyResource(t *testing.T) {
	name := fmt.Sprintf("tf-kms-key-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccKmsKeyPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckKmsKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKmsKeyConfig(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKmsKeyExists(name),
					resource.TestCheckResourceAttr(kmsKeyTestResourceName, "name", name),
					resource.TestCheckResourceAttr(kmsKeyTestResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(kmsKeyTestResourceName, "key_id"),
					resource.TestCheckResourceAttrSet(kmsKeyTestResourceName, "created_at"),
				),
			},
			{
				Config: testAccKmsKeyConfig(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKmsKeyExists(name),
					resource.TestCheckResourceAttr(kmsKeyTestResourceName, "enabled", "false"),
				),
			},
			{
				Config: testAccKmsKeyConfig(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKmsKeyExists(name),
					resource.TestCheckResourceAttr(kmsKeyTestResourceName, "enabled", "true"),
				),
			},
			{
				ResourceName:  kmsKeyTestResourceName,
				ImportState:   true,
				ImportStateId: name,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKmsKeyExists(name),
				),
			},
		},
	})
}

// testAccKmsKeyPreCheck verifies the environment and that the running RustFS
// backend supports KMS key creation. The Local dev backend does; the Static
// (read-only) backend does not. When the running backend cannot create keys
// the whole test skips gracefully instead of failing.
func testAccKmsKeyPreCheck(t *testing.T) {
	testAccPreCheck(t)

	client := newKmsKeyAccClient()
	//#nosec G117 — the probe key name is a public identifier, not a secret
	probeName := fmt.Sprintf("tf-kms-precheck-%d", acctest.RandInt())
	probeKey, err := client.CreateKmsKey(probeName)
	if err != nil {
		t.Skipf("KMS key creation is not supported on the running backend, skipping: %v", err)
	}
	// Clean up the probe key (scheduled deletion; cancellable server-side).
	if delErr := client.DeleteKmsKey(probeKey.KeyID); delErr != nil {
		t.Logf("warning: failed to clean up probe key %s: %v", probeKey.KeyID, delErr)
	}
}

func testAccKmsKeyConfig(name string, enabled bool) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_kms_key" "test" {
  name         = %q
  enabled      = %t
  skip_destroy = false
}
`, name, enabled)
}

func testAccCheckKmsKeyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[kmsKeyTestResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", kmsKeyTestResourceName)
		}
		if rs.Primary.Attributes["name"] != name {
			return fmt.Errorf("wrong name in state: %s", rs.Primary.Attributes["name"])
		}
		keyID := rs.Primary.Attributes["key_id"]
		if keyID == "" {
			return fmt.Errorf("no key_id set in state")
		}

		key, err := newKmsKeyAccClient().DescribeKmsKey(keyID)
		if err != nil {
			return fmt.Errorf("error describing KMS key %s: %s", keyID, err)
		}
		if key.KeyID != keyID {
			return fmt.Errorf("described key id mismatch: got %s, want %s", key.KeyID, keyID)
		}
		return nil
	}
}

func testAccCheckKmsKeyDestroy(s *terraform.State) error {
	client := newKmsKeyAccClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_kms_key" {
			continue
		}
		keyID := rs.Primary.Attributes["key_id"]
		if keyID == "" {
			continue
		}
		key, err := client.DescribeKmsKey(keyID)
		if err != nil {
			// Key gone (or describe refused): that satisfies destruction.
			continue
		}
		if key.KeyState != "PendingDeletion" {
			return fmt.Errorf("KMS key %s still present in state %s", keyID, key.KeyState)
		}
	}
	return nil
}

func newKmsKeyAccClient() *rustfs.RustfsAdmin {
	client := rustfs.New(&rustfs.RustfsAdminConfig{
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
	})
	return &client
}
