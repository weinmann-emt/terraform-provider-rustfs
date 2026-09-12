package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &bucketTagsRessource{}
	_ resource.ResourceWithImportState = &bucketTagsRessource{}
)

// NewBucketTagsRessource is a helper function to simplify the provider implementation.
func NewBucketTagsRessource() resource.Resource {
	return &bucketTagsRessource{}
}

// bucketTagsRessource is the resource implementation.
type bucketTagsRessource struct {
	client *AllClient
}

type bucketTagsModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Id     types.String `tfsdk:"id"`
	Tags   types.Map    `tfsdk:"tags"`
}

// Metadata returns the resource type name.
func (r *bucketTagsRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_tags"
}

// Schema defines the schema for the resource.
func (r *bucketTagsRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage the tags on an S3 bucket in rustfs",
		MarkdownDescription: "Manage the tags on an S3 bucket in rustfs",
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
			"tags": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Map of key/value tag pairs to apply to the bucket",
			},
		},
	}
}

func (r *bucketTagsRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *bucketTagsRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketTagsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagMap := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.SetBucketTagging(plan.Bucket.ValueString(), tagMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket tags",
			"Could not set bucket tagging: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "created a bucket tags resource")

	plan.Id = types.StringValue(plan.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *bucketTagsRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketTagsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagMap, err := r.client.RustClient.GetBucketTagging(state.Bucket.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchTagSet") ||
			strings.Contains(err.Error(), "NoSuchBucket") ||
			strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading bucket tags",
			"Could not read bucket tagging: "+err.Error(),
		)
		return
	}

	tags, diags := types.MapValueFrom(ctx, types.StringType, tagMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Tags = tags
	state.Id = types.StringValue(state.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *bucketTagsRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketTagsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagMap := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.SetBucketTagging(plan.Bucket.ValueString(), tagMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket tags",
			"Could not set bucket tagging: "+err.Error(),
		)
		return
	}

	plan.Id = types.StringValue(plan.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *bucketTagsRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketTagsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.RemoveBucketTagging(data.Bucket.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchTagSet") ||
			strings.Contains(err.Error(), "NoSuchBucket") ||
			strings.Contains(err.Error(), "404") {
			// Already deleted
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting bucket tags",
			"Could not remove bucket tagging: "+err.Error(),
		)
	}
}

func (r *bucketTagsRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}
