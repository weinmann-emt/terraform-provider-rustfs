package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var (
	_ resource.Resource                = &BucketDurabilityRessource{}
	_ resource.ResourceWithImportState = &BucketDurabilityRessource{}
)

func NewBucketDurabilityRessource() resource.Resource {
	return &BucketDurabilityRessource{}
}

type BucketDurabilityRessource struct {
	client *AllClient
}

type BucketDurabilityRessourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Mode   types.String `tfsdk:"mode"`
}

func (r *BucketDurabilityRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_durability"
}

func (r *BucketDurabilityRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage per-bucket durability settings in rustfs",
		MarkdownDescription: "Manage the per-bucket durability override in rustfs",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "Durability override for the bucket: strict, relaxed or none. When unset the bucket inherits the process-wide durability mode.",
				Validators: []validator.String{
					stringvalidator.OneOf("strict", "relaxed", "none"),
				},
			},
		},
	}
}

func (r *BucketDurabilityRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketDurabilityRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketDurabilityRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	d := rustfs.BucketDurability{Bucket: plan.Bucket.ValueString(), Mode: plan.Mode.ValueString()}
	read, err := r.client.RustClient.SetBucketDurability(plan.Bucket.ValueString(), d)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket durability",
			"Could not set bucket durability, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Mode = types.StringValue(read.Mode)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketDurabilityRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketDurabilityRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	read, err := r.client.RustClient.GetBucketDurability(state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket durability",
			"Could not read bucket durability, unexpected error: "+err.Error(),
		)
		return
	}
	state.Mode = types.StringValue(read.Mode)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketDurabilityRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketDurabilityRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	d := rustfs.BucketDurability{Bucket: plan.Bucket.ValueString(), Mode: plan.Mode.ValueString()}
	read, err := r.client.RustClient.SetBucketDurability(plan.Bucket.ValueString(), d)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket durability",
			"Could not update bucket durability, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Mode = types.StringValue(read.Mode)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketDurabilityRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BucketDurabilityRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RustClient.DeleteBucketDurability(data.Bucket.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bucket durability",
			"Could not delete bucket durability, unexpected error: "+err.Error(),
		)
	}
}

func (r *BucketDurabilityRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}
