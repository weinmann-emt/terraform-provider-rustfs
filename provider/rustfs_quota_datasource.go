package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &quotaDataSource{}
)

// NewQuotaDataSource is a helper function to simplify the provider implementation.
func NewQuotaDataSource() datasource.DataSource {
	return &quotaDataSource{}
}

// quotaDataSource is the data source implementation.
type quotaDataSource struct {
	client *AllClient
}

// quotaDataSourceModel maps the data source schema data.
type quotaDataSourceModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	Quota     types.Int64  `tfsdk:"quota"`
	QuotaType types.String `tfsdk:"quota_type"`
}

// Metadata returns the data source type name.
func (d *quotaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota"
}

// Schema defines the schema for the data source.
func (d *quotaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Read the quota of a RustFS bucket",
		MarkdownDescription: "Read the quota of a RustFS bucket",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket",
			},
			"quota": schema.Int64Attribute{
				Computed:    true,
				Description: "Bytes of the quota",
			},
			"quota_type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of the quota, e.g. HARD",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *quotaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the data source state with the latest data.
func (d *quotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config quotaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	quota, err := quotaReadWithRetry(ctx, config.Bucket.ValueString(), d.client.RustClient.ReadQuota)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket quota",
			"Could not read bucket quota, unexpected error: "+err.Error(),
		)
		return
	}

	config.Quota = types.Int64Value(int64(quota.Quota))
	config.QuotaType = types.StringValue(quota.Quota_Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
