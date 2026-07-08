package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &quotaRessource{}
	_ resource.ResourceWithImportState = &quotaRessource{}
)

// NewquotaRessource is a helper function to simplify the provider implementation.
func NewquotaRessource() resource.Resource {
	return &quotaRessource{}
}

// quotaRessource is the resource implementation.
type quotaRessource struct {
	client *AllClient
}

type quotaRessourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Quota  types.Int64  `tfsdk:"quota"`
}

// Metadata returns the resource type name.
func (r *quotaRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota"
}

// Schema defines the schema for the resource.
func (r *quotaRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage buckets quota in rustfs",
		MarkdownDescription: "Manage bucket quota in rustfs",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket",
			},
			"quota": schema.Int64Attribute{
				Required:    true,
				Description: "Bytes of the quota",
			},
		},
	}
}

func (r *quotaRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AllClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *quotaRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan quotaRessourceModel
	diags := req.Plan.Get(ctx, &plan)
	// ToDo: Check if bucket exists
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	q := rustfs.Quota{Bucket: plan.Bucket.ValueString(), Quota: int(plan.Quota.ValueInt64()), Quota_Type: "HARD"}
	_, err := r.client.RustClient.SetQuota(q)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket quota",
			"Could not create bucket quota, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created a resource")
	// plan.ID = types.StringValue(account.AccessKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *quotaRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state quotaRessourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Read
	read, _ := r.client.RustClient.ReadQuota(state.Bucket.ValueString())
	// Save update status
	state.Quota = types.Int64Value(int64(read.Quota))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *quotaRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan quotaRessourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := rustfs.Quota{Bucket: plan.Bucket.ValueString(), Quota: int(plan.Quota.ValueInt64()), Quota_Type: "HARD"}
	read, err := r.client.RustClient.SetQuota(q)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket quota",
			"Could not update bucket quota, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Quota = types.Int64Value(int64(read.Quota))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *quotaRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data quotaRessourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.DeletQuota(data.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bucket quota",
			"Could not delete bucket quota, unexpected error: "+err.Error(),
		)
	}
}

func (r *quotaRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}
