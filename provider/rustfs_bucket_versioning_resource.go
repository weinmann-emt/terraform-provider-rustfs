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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/minio/minio-go/v7"
)

var (
	_ resource.Resource                = &BucketVersioningResource{}
	_ resource.ResourceWithImportState = &BucketVersioningResource{}
)

type BucketVersioningResource struct {
	client *AllClient
}

type BucketVersioningResourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Status types.String `tfsdk:"status"`
}

func NewBucketVersioningResource() resource.Resource {
	return &BucketVersioningResource{}
}

func (r *BucketVersioningResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_versioning"
}

func (r *BucketVersioningResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS bucket versioning",
		MarkdownDescription: "Manage RustFS bucket versioning configuration",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Required:    true,
				Description: "Versioning status: Enabled or Suspended.",
			},
		},
	}
}

func (r *BucketVersioningResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketVersioningResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketVersioningResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Minio.SetBucketVersioning(ctx, plan.Bucket.ValueString(), minio.BucketVersioningConfiguration{
		Status: plan.Status.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting bucket versioning",
			"Could not set bucket versioning: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "created bucket versioning")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketVersioningResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketVersioningResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.Minio.GetBucketVersioning(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket versioning",
			"Could not read bucket versioning: "+err.Error(),
		)
		return
	}

	state.Status = types.StringValue(config.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketVersioningResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketVersioningResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Minio.SetBucketVersioning(ctx, plan.Bucket.ValueString(), minio.BucketVersioningConfiguration{
		Status: plan.Status.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket versioning",
			"Could not update bucket versioning: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketVersioningResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BucketVersioningResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Minio.SuspendVersioning(ctx, data.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error suspending bucket versioning",
			"Could not suspend bucket versioning: "+err.Error(),
		)
		return
	}
}

func (r *BucketVersioningResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}
