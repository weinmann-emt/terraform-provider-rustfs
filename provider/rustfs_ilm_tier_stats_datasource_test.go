package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestIlmTierStatsDataSourceSchema(t *testing.T) {
	d := NewIlmTierStatsDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "tiers"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing schema attribute %q", attr)
		}
	}
	nested, ok := resp.Schema.Attributes["tiers"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("tiers attribute is not a ListNestedAttribute")
	}
	for _, attr := range []string{"name", "num_objects", "num_versions", "total_size"} {
		if _, ok := nested.NestedObject.Attributes[attr]; !ok {
			t.Errorf("missing nested schema attribute %q", attr)
		}
	}
}

func TestAccIlmTierStatsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "rustfs_ilm_tier_stats" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rustfs_ilm_tier_stats.test", "id"),
				),
			},
		},
	})
}
