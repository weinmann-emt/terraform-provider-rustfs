package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestBucketObjectLockResourceSchema(t *testing.T) {
	r := NewBucketObjectLockResource()
	resp := &resource.SchemaResponse{}
	r.Schema(nil, resource.SchemaRequest{}, resp)

	if diags := resp.Diagnostics; diags.HasError() {
		t.Fatalf("schema diagnostics: %v", diags)
	}

	attrs := resp.Schema.GetAttributes()
	if _, ok := attrs["bucket"]; !ok {
		t.Error("expected bucket attribute")
	}
	if _, ok := attrs["mode"]; !ok {
		t.Error("expected mode attribute")
	}
	if _, ok := attrs["days"]; !ok {
		t.Error("expected days attribute")
	}
	if _, ok := attrs["years"]; !ok {
		t.Error("expected years attribute")
	}
}

func TestBucketObjectLockResourceMetadata(t *testing.T) {
	r := NewBucketObjectLockResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(nil, resource.MetadataRequest{ProviderTypeName: "rustfs"}, resp)

	if resp.TypeName != "rustfs_bucket_object_lock" {
		t.Errorf("expected rustfs_bucket_object_lock, got %s", resp.TypeName)
	}
}
