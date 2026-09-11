package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserMfaDataSource{}

type UserMfaDataSource struct {
	client *AllClient
}

type UserMfaDataSourceModel struct {
	AccessKey              types.String `tfsdk:"access_key"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	ActivatedAt            types.String `tfsdk:"activated_at"`
	RecoveryCodesRemaining types.Int64  `tfsdk:"recovery_codes_remaining"`
}

func NewUserMfaDataSource() datasource.DataSource {
	return &UserMfaDataSource{}
}

func (d *UserMfaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_mfa"
}

func (d *UserMfaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Read the MFA status of a RustFS user",
		MarkdownDescription: "Fetch the second-factor (MFA) status of a RustFS user by access key",
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Required:    true,
				Description: "Access key of the user.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether two-factor authentication is enabled for the user.",
			},
			"activated_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC3339 timestamp of when the second factor was activated, if enabled.",
			},
			"recovery_codes_remaining": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of unused recovery codes remaining for the user.",
			},
		},
	}
}

func (d *UserMfaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserMfaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserMfaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := d.client.RustClient.ReadUserMFA(config.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MFA status",
			"Could not read MFA status: "+err.Error(),
		)
		return
	}

	config.Enabled = types.BoolValue(status.Enabled)
	config.ActivatedAt = types.StringValue(status.ActivatedAt)
	config.RecoveryCodesRemaining = types.Int64Value(int64(status.RecoveryCodesRemaining))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
