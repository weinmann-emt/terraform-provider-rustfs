package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUserMfaDataSourceSchema(t *testing.T) {
	d := NewUserMfaDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.TODO(), datasource.SchemaRequest{}, resp)

	if diags := resp.Diagnostics; diags.HasError() {
		t.Fatalf("schema diagnostics: %v", diags)
	}

	attrs := resp.Schema.GetAttributes()
	for _, want := range []string{"access_key", "enabled", "activated_at", "recovery_codes_remaining"} {
		if _, ok := attrs[want]; !ok {
			t.Errorf("expected %s attribute", want)
		}
	}
}

func TestUserMfaDataSourceMetadata(t *testing.T) {
	d := NewUserMfaDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.TODO(), datasource.MetadataRequest{ProviderTypeName: "rustfs"}, resp)

	if resp.TypeName != "rustfs_user_mfa" {
		t.Errorf("expected rustfs_user_mfa, got %s", resp.TypeName)
	}
}

func TestAccUserMfaDataSource(t *testing.T) {
	name := fmt.Sprintf("tf-test-mfa-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_user" "test" {
  name       = "%s"
  access_key = "%s"
  secret_key = "superSecret123!"
  status     = "enabled"
  policy     = ""
}

data "rustfs_user_mfa" "test" {
  access_key = rustfs_user.test.access_key
}
`, name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.rustfs_user_mfa.test", "access_key", name),
					resource.TestCheckResourceAttrSet("data.rustfs_user_mfa.test", "enabled"),
				),
			},
		},
	})
}
