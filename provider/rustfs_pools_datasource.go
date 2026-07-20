package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PoolsDataSource{}

type PoolsDataSource struct {
	client *AllClient
}

type PoolsDataSourceModel struct {
	Names types.List `tfsdk:"names"`
}

func NewPoolsDataSource() datasource.DataSource {
	return &PoolsDataSource{}
}

func (d *PoolsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pools"
}

func (d *PoolsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "List RustFS storage pools",
		MarkdownDescription: "List all RustFS storage pools and their status",
		Attributes: map[string]schema.Attribute{
			"names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of storage pool names.",
			},
		},
	}
}

func (d *PoolsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AllClient, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *PoolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PoolsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pools, err := d.client.RustClient.ListPools()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing pools",
			"Could not list pools: "+err.Error(),
		)
		return
	}

	var names []string
	for _, p := range pools {
		names = append(names, p.Name)
	}

	poolNames, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	config.Names = poolNames
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
