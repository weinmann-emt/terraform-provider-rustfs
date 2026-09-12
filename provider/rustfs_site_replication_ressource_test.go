package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSiteReplicationRessourceSchema(t *testing.T) {
	r := NewSiteReplicationRessource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"name", "endpoint", "access_key", "secret_key", "skip_tls_verify", "ca_cert_pem"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing schema attribute %q", attr)
		}
	}
}
