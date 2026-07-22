package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &BucketMetadataBackupImportResource{}

type BucketMetadataBackupImportResource struct {
	client *AllClient
}

type BucketMetadataBackupImportResourceModel struct {
	ContentBase64 types.String `tfsdk:"content_base64"`
}

func NewBucketMetadataBackupImportResource() resource.Resource {
	return &BucketMetadataBackupImportResource{}
}

func (r *BucketMetadataBackupImportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_metadata_backup_import"
}

func (r *BucketMetadataBackupImportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Import bucket metadata from a ZIP archive",
		Attributes: map[string]schema.Attribute{
			"content_base64": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Base64-encoded ZIP archive with bucket metadata.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *BucketMetadataBackupImportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *AllClient, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *BucketMetadataBackupImportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketMetadataBackupImportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := base64.StdEncoding.DecodeString(plan.ContentBase64.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid base64 data", err.Error())
		return
	}

	if err := r.client.RustClient.ImportBucketMetadata(data); err != nil {
		resp.Diagnostics.AddError("Error importing bucket metadata", "Could not import: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketMetadataBackupImportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketMetadataBackupImportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketMetadataBackupImportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketMetadataBackupImportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := base64.StdEncoding.DecodeString(plan.ContentBase64.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid base64 data", err.Error())
		return
	}

	if err := r.client.RustClient.ImportBucketMetadata(data); err != nil {
		resp.Diagnostics.AddError("Error importing bucket metadata", "Could not import: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketMetadataBackupImportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
