package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/minio/minio-go/v7"
)

var (
	_ resource.Resource                = &BucketObjectLockResource{}
	_ resource.ResourceWithImportState = &BucketObjectLockResource{}
)

type BucketObjectLockResource struct {
	client *AllClient
}

type BucketObjectLockResourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Mode   types.String `tfsdk:"mode"`
	Days   types.Int64  `tfsdk:"days"`
	Years  types.Int64  `tfsdk:"years"`
}

func NewBucketObjectLockResource() resource.Resource {
	return &BucketObjectLockResource{}
}

func (r *BucketObjectLockResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_object_lock"
}

func (r *BucketObjectLockResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS bucket object lock",
		MarkdownDescription: "Manage RustFS bucket object lock configuration",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket. Bucket must have been created with object lock enabled.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "Object lock retention mode: COMPLIANCE or GOVERNANCE.",
			},
			"days": schema.Int64Attribute{
				Optional:    true,
				Description: "Retention period in days.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"years": schema.Int64Attribute{
				Optional:    true,
				Description: "Retention period in years.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *BucketObjectLockResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketObjectLockResource) setConfig(ctx context.Context, plan BucketObjectLockResourceModel) error {
	mode := minio.RetentionMode(plan.Mode.ValueString())
	daysVal := plan.Days.ValueInt64()
	yearsVal := plan.Years.ValueInt64()
	if daysVal < 0 {
		daysVal = 0
	}
	if yearsVal < 0 {
		yearsVal = 0
	}
	days := uint(daysVal)
	years := uint(yearsVal)
	var validity *uint
	var unit *minio.ValidityUnit

	if days > 0 {
		validity = &days
		d := minio.Days
		unit = &d
	} else if years > 0 {
		validity = &years
		y := minio.Years
		unit = &y
	}

	return r.client.Minio.SetObjectLockConfig(ctx, plan.Bucket.ValueString(), &mode, validity, unit)
}

func (r *BucketObjectLockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketObjectLockResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.setConfig(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error setting object lock",
			"Could not set object lock: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketObjectLockResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketObjectLockResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mode, validity, unit, err := r.client.Minio.GetBucketObjectLockConfig(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading object lock",
			"Could not read object lock: "+err.Error(),
		)
		return
	}

	if mode != nil {
		state.Mode = types.StringValue(string(*mode))
	}
	if validity != nil && unit != nil {
		switch *unit {
		case minio.Days:
			state.Days = types.Int64Value(int64(*validity)) // #nosec G115
		case minio.Years:
			state.Years = types.Int64Value(int64(*validity)) // #nosec G115
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketObjectLockResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketObjectLockResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.setConfig(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error updating object lock",
			"Could not update object lock: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketObjectLockResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Object lock cannot be removed from a bucket; only the bucket itself can be deleted.
}

func (r *BucketObjectLockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}
