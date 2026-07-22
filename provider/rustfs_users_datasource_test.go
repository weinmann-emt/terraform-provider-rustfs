package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestUsersDataSourceSchema(t *testing.T) {
	d := NewUsersDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(nil, datasource.SchemaRequest{}, resp)

	if diags := resp.Diagnostics; diags.HasError() {
		t.Fatalf("schema diagnostics: %v", diags)
	}

	attrs := resp.Schema.GetAttributes()
	if _, ok := attrs["bucket"]; !ok {
		t.Error("expected bucket attribute")
	}
	if _, ok := attrs["access_keys"]; !ok {
		t.Error("expected access_keys attribute")
	}
}

func TestUsersDataSourceMetadata(t *testing.T) {
	d := NewUsersDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(nil, datasource.MetadataRequest{ProviderTypeName: "rustfs"}, resp)

	if resp.TypeName != "rustfs_users" {
		t.Errorf("expected rustfs_users, got %s", resp.TypeName)
	}
}
