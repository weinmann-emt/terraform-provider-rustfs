package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &GroupsDataSource{}

type GroupsDataSource struct {
	client *AllClient
}

type GroupsDataSourceModel struct {
	Groups types.Set `tfsdk:"groups"`
}

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

func (d *GroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all IAM groups for RustFS",
		MarkdownDescription: "Lists all IAM groups for RustFS",
		Attributes: map[string]schema.Attribute{
			"groups": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				Description:         "Set of IAM group names",
				MarkdownDescription: "Set of IAM group names",
			},
		},
	}
}

func (d *GroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	names, err := d.client.RustClient.ListGroups()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing groups",
			"Could not list groups: "+err.Error(),
		)
		return
	}
	set, diags := types.SetValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := GroupsDataSourceModel{
		Groups: set,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
