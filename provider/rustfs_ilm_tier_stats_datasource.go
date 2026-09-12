package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IlmTierStatsDataSource{}

type IlmTierStatsDataSource struct {
	client *AllClient
}

type IlmTierStatsDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Tiers types.List   `tfsdk:"tiers"`
}

type TierStatsModel struct {
	Name        types.String `tfsdk:"name"`
	NumObjects  types.Int64  `tfsdk:"num_objects"`
	NumVersions types.Int64  `tfsdk:"num_versions"`
	TotalSize   types.Int64  `tfsdk:"total_size"`
}

func NewIlmTierStatsDataSource() datasource.DataSource {
	return &IlmTierStatsDataSource{}
}

func (d *IlmTierStatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ilm_tier_stats"
}

func (d *IlmTierStatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Expose per-tier ILM storage statistics",
		MarkdownDescription: "Exposes per-tier ILM storage statistics (object counts and sizes) from RustFS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"tiers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Per-tier statistics.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Tier name.",
						},
						"num_objects": schema.Int64Attribute{
							Computed:    true,
							Description: "Number of objects on the tier.",
						},
						"num_versions": schema.Int64Attribute{
							Computed:    true,
							Description: "Number of object versions on the tier.",
						},
						"total_size": schema.Int64Attribute{
							Computed:    true,
							Description: "Total object size in bytes.",
						},
					},
				},
			},
		},
	}
}

func (d *IlmTierStatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AllClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *IlmTierStatsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IlmTierStatsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stats, err := d.client.RustClient.TierStats()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tier stats",
			"Could not read tier stats: "+err.Error(),
		)
		return
	}

	names := make([]string, 0, len(stats))
	for name := range stats {
		names = append(names, name)
	}
	sort.Strings(names)

	tiers := make([]TierStatsModel, 0, len(names))
	for _, name := range names {
		stat := stats[name]
		tiers = append(tiers, TierStatsModel{
			Name:        types.StringValue(name),
			NumObjects:  types.Int64Value(stat.NumObjects),
			NumVersions: types.Int64Value(stat.NumVersions),
			TotalSize:   types.Int64Value(stat.TotalSize),
		})
	}

	elementType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":         types.StringType,
		"num_objects":  types.Int64Type,
		"num_versions": types.Int64Type,
		"total_size":   types.Int64Type,
	}}
	tiersValue, diags := types.ListValueFrom(ctx, elementType, tiers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue("tier-stats")
	config.Tiers = tiersValue
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
