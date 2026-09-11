package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

const auditTargetResourceName = "rustfs_audit_target.test"

func TestAccAuditTargetResource(t *testing.T) {
	name := fmt.Sprintf("tf-audit-%d", acctest.RandInt())
	endpoint := fmt.Sprintf("https://hooks.example.com/webhook/%s", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccAuditTargetPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAuditTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuditTargetConfig(name, endpoint),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAuditTargetExists(name),
					resource.TestCheckResourceAttr(auditTargetResourceName, "target_type", "audit_webhook"),
					resource.TestCheckResourceAttr(auditTargetResourceName, "target_name", name),
					resource.TestCheckResourceAttr(auditTargetResourceName, "endpoint", endpoint),
					resource.TestCheckResourceAttr(auditTargetResourceName, "auth_token", "superSecret"),
					resource.TestCheckResourceAttr(auditTargetResourceName, "comment", "managed by terraform"),
					resource.TestCheckResourceAttrSet(auditTargetResourceName, "health_state"),
				),
			},
			{
				Config: testAccAuditTargetConfigUpdated(name, endpoint),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAuditTargetExists(name),
					resource.TestCheckResourceAttr(auditTargetResourceName, "comment", "updated by terraform"),
					resource.TestCheckResourceAttrSet(auditTargetResourceName, "health_state"),
				),
			},
			{
				ResourceName:      auditTargetResourceName,
				ImportState:       true,
				ImportStateId:     "audit_webhook/" + name,
				ImportStateVerify: true,
				// The admin list endpoint does not echo config values, so these
				// attributes cannot be verified after import.
				ImportStateVerifyIdentifierAttribute: "target_name",
				ImportStateVerifyIgnore:              []string{"endpoint", "auth_token", "comment", "queue_limit", "queue_dir", "client_cert", "client_key", "client_ca", "skip_tls_verify", "health_state", "health_reason", "status"},
			},
		},
	})
}

// testAccAuditTargetPreCheck verifies the environment and that the running
// RustFS has the audit module enabled (the admin API rejects target management
// otherwise). The list endpoint succeeds even when the module is disabled, so
// the check probes an actual create and skips when the server reports the
// audit module is disabled.
func testAccAuditTargetPreCheck(t *testing.T) {
	testAccPreCheck(t)

	client := testAccRustClient()
	probe := fmt.Sprintf("tf-audit-probe-%d", acctest.RandInt())
	probeConfig := []rustfs.AuditTargetKeyValue{
		{Key: "endpoint", Value: "https://hooks.example.com/webhook/" + probe},
	}
	err := client.SetAuditTarget("audit_webhook", probe, probeConfig)
	if err != nil {
		if strings.Contains(err.Error(), "audit module is disabled") {
			t.Skipf("audit module is disabled on the running RustFS; skipping: %v", err)
		}
		t.Fatalf("audit target management is not available on the running RustFS, skipping: %v", err)
	}
	// Clean up the probe target so it does not linger in the server config.
	if err := client.ResetAuditTarget("audit_webhook", probe); err != nil {
		t.Logf("could not clean up audit target probe %s: %v", probe, err)
	}
}

func testAccAuditTargetConfig(name, endpoint string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_audit_target" "test" {
  target_type = "audit_webhook"
  target_name = "%s"
  endpoint    = "%s"
  auth_token  = "superSecret"
  comment     = "managed by terraform"
}
`, name, endpoint)
}

func testAccAuditTargetConfigUpdated(name, endpoint string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_audit_target" "test" {
  target_type = "audit_webhook"
  target_name = "%s"
  endpoint    = "%s"
  auth_token  = "superSecret"
  comment     = "updated by terraform"
}
`, name, endpoint)
}

func testAccCheckAuditTargetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[auditTargetResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", auditTargetResourceName)
		}
		targetName := rs.Primary.Attributes["target_name"]
		if targetName == "" {
			return fmt.Errorf("no target_name set")
		}

		client := testAccRustClient()
		targets, err := client.ListAuditTargets()
		if err != nil {
			return err
		}
		for _, target := range targets {
			if target.AccountID == name && target.Service == "webhook" {
				return nil
			}
		}
		return fmt.Errorf("audit target %s not found in list", name)
	}
}

func testAccCheckAuditTargetDestroy(s *terraform.State) error {
	client := testAccRustClient()
	targets, err := client.ListAuditTargets()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rustfs_audit_target" {
			continue
		}
		name := rs.Primary.Attributes["target_name"]
		for _, target := range targets {
			if strings.EqualFold(target.AccountID, name) {
				return fmt.Errorf("audit target %s still exists", name)
			}
		}
	}
	return nil
}
