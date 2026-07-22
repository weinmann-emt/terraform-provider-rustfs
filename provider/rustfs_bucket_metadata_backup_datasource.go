package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &BucketMetadataBackupDataSource{}

type BucketMetadataBackupDataSource struct {
	client *AllClient
}

type BucketMetadataBackupDataSourceModel struct {
	ContentBase64 types.String `tfsdk:"content_base64"`
}

func NewBucketMetadataBackupDataSource() datasource.DataSource {
	return &BucketMetadataBackupDataSource{}
}

func (d *BucketMetadataBackupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_metadata_backup"
}

func (d *BucketMetadataBackupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Export bucket metadata as a ZIP archive",
		Attributes: map[string]schema.Attribute{
			"content_base64": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Base64-encoded ZIP archive with bucket metadata.",
			},
		},
	}
}

func (d *BucketMetadataBackupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *AllClient, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *BucketMetadataBackupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BucketMetadataBackupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := d.client.RustClient.ExportBucketMetadata()
	if err != nil {
		resp.Diagnostics.AddError("Error exporting bucket metadata", "Could not export: "+err.Error())
		return
	}

	config.ContentBase64 = types.StringValue(base64.StdEncoding.EncodeToString(data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
