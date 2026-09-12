package provider

import (
	"context"
	"os"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
) // originalModuleSwitch captures the server's module switch state before the
// test mutates it, so the test can restore the original state afterwards.
var originalModuleSwitch *rustfs.ModuleSwitchState

func readModuleSwitchState(t *testing.T) *rustfs.ModuleSwitchState {
	t.Helper()
	client := rustfs.New(&rustfs.RustfsAdminConfig{
		AccessKey:    os.Getenv("RUSTFS_USER"),
		AccessSecret: os.Getenv("RUSTFS_SECRET"),
		Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
	})
	state, err := client.GetModuleSwitches()
	if err != nil {
		t.Fatalf("failed to read original module switches: %v", err)
	}
	return state
}

func TestAccModuleSwitchResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			originalModuleSwitch = readModuleSwitchState(t)
			t.Cleanup(func() {
				client := rustfs.New(&rustfs.RustfsAdminConfig{
					AccessKey:    os.Getenv("RUSTFS_USER"),
					AccessSecret: os.Getenv("RUSTFS_SECRET"),
					Endpoint:     os.Getenv("RUSTFS_ENDPOINT"),
				})
				update := rustfs.ModuleSwitchUpdate{
					NotifyEnabled: originalModuleSwitch.NotifyEnabled,
					AuditEnabled:  originalModuleSwitch.AuditEnabled,
				}
				if _, err := client.SetModuleSwitches(update); err != nil {
					t.Errorf("failed to restore original module switches: %v", err)
				}
			})
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "rustfs_module_switch" "test" {
  notify_enabled = true
  audit_enabled  = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_module_switch.test", "id", "module-switches"),
					resource.TestCheckResourceAttr("rustfs_module_switch.test", "notify_enabled", "true"),
					resource.TestCheckResourceAttr("rustfs_module_switch.test", "audit_enabled", "true"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "rustfs_module_switch" "test" {
  notify_enabled = false
  audit_enabled  = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rustfs_module_switch.test", "notify_enabled", "false"),
					resource.TestCheckResourceAttr("rustfs_module_switch.test", "audit_enabled", "true"),
				),
			},
		},
	})
}

func TestModuleSwitchResourceSchema(t *testing.T) {
	r := NewModuleSwitchRessource()
	resp := &fwresource.SchemaResponse{}
	r.Schema(context.TODO(), fwresource.SchemaRequest{}, resp)

	if diags := resp.Diagnostics; diags.HasError() {
		t.Fatalf("schema diagnostics: %v", diags)
	}

	attrs := resp.Schema.GetAttributes()
	for _, name := range []string{"id", "notify_enabled", "audit_enabled"} {
		if _, ok := attrs[name]; !ok {
			t.Errorf("expected %s attribute", name)
		}
	}
}

func TestModuleSwitchUpdateFromModel(t *testing.T) {
	update := moduleSwitchUpdateFromModel(ModuleSwitchRessourceModel{
		NotifyEnabled: types.BoolValue(true),
		AuditEnabled:  types.BoolValue(false),
	})
	if !update.NotifyEnabled || update.AuditEnabled {
		t.Fatalf("unexpected update: %+v", update)
	}
}
