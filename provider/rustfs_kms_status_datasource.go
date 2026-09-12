package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &kmsStatusDataSource{}
)

// NewKmsStatusDataSource is a helper function to simplify the provider
// implementation.
func NewKmsStatusDataSource() datasource.DataSource {
	return &kmsStatusDataSource{}
}

// kmsStatusDataSource is the data source implementation.
type kmsStatusDataSource struct {
	client *AllClient
}

// kmsStatusDataSourceModel maps the data source schema to the combined KMS
// status and configuration documents.
type kmsStatusDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	BackendType     types.String `tfsdk:"backend_type"`
	BackendStatus   types.String `tfsdk:"backend_status"`
	DefaultKeyID    types.String `tfsdk:"default_key_id"`
	CacheEnabled    types.Bool   `tfsdk:"cache_enabled"`
	Backend         types.String `tfsdk:"backend"`
	CacheMaxKeys    types.Int64  `tfsdk:"cache_max_keys"`
	CacheTTLSeconds types.Int64  `tfsdk:"cache_ttl_seconds"`
	StatusJSON      types.String `tfsdk:"status_json"`
	ConfigJSON      types.String `tfsdk:"config_json"`
}

// Metadata returns the data source type name.
func (d *kmsStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms_status"
}

// Schema defines the schema for the data source.
func (d *kmsStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Read the RustFS KMS status and backend configuration",
		MarkdownDescription: "Read the RustFS KMS status and backend configuration from `GET /rustfs/admin/v3/kms/status` and `GET /rustfs/admin/v3/kms/config`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Fixed identifier for the KMS status data source.",
			},
			"backend_type": schema.StringAttribute{
				Computed:    true,
				Description: "Name or type of the configured KMS backend as reported by /kms/status (e.g. local, vault-kv2, aws).",
			},
			"backend_status": schema.StringAttribute{
				Computed:    true,
				Description: "Health status of the KMS backend (healthy, unhealthy or error).",
			},
			"default_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "Key ID used when no explicit key is specified.",
			},
			"cache_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the KMS key cache is enabled.",
			},
			"backend": schema.StringAttribute{
				Computed:    true,
				Description: "Configured KMS backend as reported by /kms/config (e.g. local, vault-kv2, aws).",
			},
			"cache_max_keys": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum number of keys held in the KMS cache.",
			},
			"cache_ttl_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "Time-to-live of cached KMS entries, in seconds.",
			},
			"status_json": schema.StringAttribute{
				Computed:            true,
				Description:         "Full raw JSON document returned by GET /rustfs/admin/v3/kms/status.",
				MarkdownDescription: "Full raw JSON document returned by `GET /rustfs/admin/v3/kms/status`.",
			},
			"config_json": schema.StringAttribute{
				Computed:            true,
				Description:         "Full raw JSON document returned by GET /rustfs/admin/v3/kms/config.",
				MarkdownDescription: "Full raw JSON document returned by `GET /rustfs/admin/v3/kms/config`.",
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *kmsStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *kmsStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config kmsStatusDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := d.client.RustClient.KmsStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KMS status",
			"Could not read KMS status, unexpected error: "+err.Error(),
		)
		return
	}
	kmsConfig, err := d.client.RustClient.KmsConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading KMS config",
			"Could not read KMS config, unexpected error: "+err.Error(),
		)
		return
	}

	statusJSON, err := json.Marshal(status)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling KMS status",
			"Could not marshal KMS status: "+err.Error(),
		)
		return
	}
	configJSON, err := json.Marshal(kmsConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling KMS config",
			"Could not marshal KMS config: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue("kms-status")
	config.BackendType = types.StringValue(status.BackendType)
	config.BackendStatus = types.StringValue(status.BackendStatus)
	config.DefaultKeyID = optionalStringValue(status.DefaultKeyID)
	config.CacheEnabled = types.BoolValue(status.CacheEnabled)
	config.Backend = types.StringValue(kmsConfig.Backend)
	//#nosec G115 — KMS cache configuration values are small and fit in int64
	config.CacheMaxKeys = types.Int64Value(int64(kmsConfig.CacheMaxKeys))
	//#nosec G115 — KMS cache configuration values are small and fit in int64
	config.CacheTTLSeconds = types.Int64Value(int64(kmsConfig.CacheTTLSeconds))
	config.StatusJSON = types.StringValue(string(statusJSON))
	config.ConfigJSON = types.StringValue(string(configJSON))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func optionalStringValue(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}
