package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/minio/minio-go/v7/pkg/sse"
)

var (
	_ resource.Resource                = &BucketEncryptionResource{}
	_ resource.ResourceWithImportState = &BucketEncryptionResource{}
)

type BucketEncryptionResource struct {
	client *AllClient
}

type BucketEncryptionResourceModel struct {
	Bucket         types.String `tfsdk:"bucket"`
	Algorithm      types.String `tfsdk:"algorithm"`
	KmsMasterKeyID types.String `tfsdk:"kms_master_key_id"`
}

func NewBucketEncryptionResource() resource.Resource {
	return &BucketEncryptionResource{}
}

func (r *BucketEncryptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_encryption"
}

func (r *BucketEncryptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS bucket encryption",
		MarkdownDescription: "Manage RustFS bucket server-side encryption configuration",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm": schema.StringAttribute{
				Required:    true,
				Description: "Encryption algorithm: AES256 or aws:kms.",
			},
			"kms_master_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "KMS Master Key ID. Required when algorithm is aws:kms.",
			},
		},
	}
}

func (r *BucketEncryptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AllClient, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *BucketEncryptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketEncryptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Minio.SetBucketEncryption(ctx, plan.Bucket.ValueString(), buildEncryptionConfig(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting bucket encryption",
			"Could not set bucket encryption: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketEncryptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketEncryptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.Minio.GetBucketEncryption(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket encryption",
			"Could not read bucket encryption: "+err.Error(),
		)
		return
	}

	if len(config.Rules) > 0 {
		state.Algorithm = types.StringValue(config.Rules[0].Apply.SSEAlgorithm)
		state.KmsMasterKeyID = types.StringValue(config.Rules[0].Apply.KmsMasterKeyID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketEncryptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketEncryptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Minio.SetBucketEncryption(ctx, plan.Bucket.ValueString(), buildEncryptionConfig(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket encryption",
			"Could not update bucket encryption: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketEncryptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BucketEncryptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Minio.RemoveBucketEncryption(ctx, data.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing bucket encryption",
			"Could not remove bucket encryption: "+err.Error(),
		)
		return
	}
}

func (r *BucketEncryptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}

func buildEncryptionConfig(plan BucketEncryptionResourceModel) *sse.Configuration {
	return &sse.Configuration{
		Rules: []sse.Rule{
			{
				Apply: sse.ApplySSEByDefault{
					SSEAlgorithm:   plan.Algorithm.ValueString(),
					KmsMasterKeyID: plan.KmsMasterKeyID.ValueString(),
				},
			},
		},
	}
}
