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
)

var (
	_ resource.Resource                = &bucketPolicyRessource{}
	_ resource.ResourceWithImportState = &bucketPolicyRessource{}
)

func NewBucketPolicyRessource() resource.Resource {
	return &bucketPolicyRessource{}
}

type bucketPolicyRessource struct {
	client *AllClient
}

type bucketPolicyModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Id     types.String `tfsdk:"id"`
	Policy types.String `tfsdk:"policy"`
}

func (r *bucketPolicyRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_policy"
}

func (r *bucketPolicyRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage the raw S3 bucket policy document in rustfs",
		MarkdownDescription: "Manage the raw S3 bucket policy document in rustfs",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Name of the bucket",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The bucket name",
			},
			"policy": schema.StringAttribute{
				Required:    true,
				Description: "Raw S3 bucket policy document as JSON (see https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-policy-language-overview.html)",
			},
		},
	}
}

func (r *bucketPolicyRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketPolicyRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.SetBucketPolicy(plan.Bucket.ValueString(), plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket policy",
			"Could not set bucket policy: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "created a bucket policy resource")

	plan.Id = types.StringValue(plan.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bucketPolicyRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.RustClient.GetBucketPolicy(state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket policy",
			"Could not read bucket policy: "+err.Error(),
		)
		return
	}
	if policy == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Policy = types.StringValue(policy)
	state.Id = types.StringValue(state.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bucketPolicyRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.SetBucketPolicy(plan.Bucket.ValueString(), plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket policy",
			"Could not set bucket policy: "+err.Error(),
		)
		return
	}

	plan.Id = types.StringValue(plan.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *bucketPolicyRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.RemoveBucketPolicy(data.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bucket policy",
			"Could not remove bucket policy: "+err.Error(),
		)
	}
}

func (r *bucketPolicyRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}
