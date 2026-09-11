package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IAMPolicyDataSource{}

type IAMPolicyDataSource struct {
	client *AllClient
}

type iamPolicyStatementDataSourceModel struct {
	Effect   types.String `tfsdk:"effect"`
	Action   types.Set    `tfsdk:"action"`
	Resource types.Set    `tfsdk:"resource"`
}

type IAMPolicyDataSourceModel struct {
	Name      types.String                        `tfsdk:"name"`
	Version   types.String                        `tfsdk:"version"`
	Statement []iamPolicyStatementDataSourceModel `tfsdk:"statement"`
}

func NewIAMPolicyDataSource() datasource.DataSource {
	return &IAMPolicyDataSource{}
}

func (d *IAMPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_policy"
}

func (d *IAMPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Inspect a single canned IAM policy",
		MarkdownDescription: "Fetch the details of a single canned IAM policy by name",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the canned policy.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Policy version (2012-10-17).",
			},
			"statement": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							Computed: true,
						},
						"action": schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
						"resource": schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *IAMPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IAMPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IAMPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.RustClient.ReadPolicy(config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading policy",
			"Could not read policy: "+err.Error(),
		)
		return
	}

	config.Name = types.StringValue(policy.Name)
	config.Version = types.StringValue(policy.Version)
	config.Statement = []iamPolicyStatementDataSourceModel{}
	for _, s := range policy.Statement {
		actions, diags := types.SetValueFrom(ctx, types.StringType, s.Action)
		resp.Diagnostics.Append(diags...)
		resources, diags := types.SetValueFrom(ctx, types.StringType, s.Resource)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.Statement = append(config.Statement, iamPolicyStatementDataSourceModel{
			Effect:    types.StringValue(s.Effect),
			Action:    actions,
			Ressource: resources,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
