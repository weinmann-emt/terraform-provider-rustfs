package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestBucketDurabilityRessourceSchema(t *testing.T) {
	r := NewBucketDurabilityRessource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	attrs := resp.Schema.GetAttributes()
	for _, attr := range []string{"bucket", "mode"} {
		if _, ok := attrs[attr]; !ok {
			t.Errorf("missing schema attribute %q", attr)
		}
	}
}

func TestBucketDurabilityRessourceMetadata(t *testing.T) {
	r := NewBucketDurabilityRessource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "rustfs"}, resp)
	if resp.TypeName != "rustfs_bucket_durability" {
		t.Errorf("expected rustfs_bucket_durability, got %s", resp.TypeName)
	}
}
