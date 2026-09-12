package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &MetricsDataSource{}

type MetricsDataSource struct {
	client *AllClient
}

type MetricsDataSourceModel struct {
	Metrics types.String `tfsdk:"metrics"`
}

func NewMetricsDataSource() datasource.DataSource {
	return &MetricsDataSource{}
}

func (d *MetricsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics"
}

func (d *MetricsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Metrics stream for RustFS server",
		MarkdownDescription: "Metrics stream for RustFS server",
		Attributes: map[string]schema.Attribute{
			"metrics": schema.StringAttribute{
				Computed:            true,
				Description:         "Raw metrics stream from the /metrics endpoint",
				MarkdownDescription: "Raw metrics stream from the /metrics endpoint",
			},
		},
	}
}

func (d *MetricsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MetricsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	raw, err := d.client.RustClient.GetMetrics()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading metrics",
			"Could not read metrics, unexpected error: "+err.Error(),
		)
		return
	}
	state := MetricsDataSourceModel{
		Metrics: types.StringValue(raw),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
