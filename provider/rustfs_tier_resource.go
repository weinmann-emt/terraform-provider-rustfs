package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &TierResource{}
	_ resource.ResourceWithImportState = &TierResource{}
)

type TierResource struct {
	client *AllClient
}

type tierResourceModel struct {
	Name       types.String `tfsdk:"name"`
	TierType   types.String `tfsdk:"tier_type"`
	ConfigJson types.String `tfsdk:"config_json"`
}

func NewTierResource() resource.Resource {
	return &TierResource{}
}

func (r *TierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tier"
}

func (r *TierResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS storage tiers",
		MarkdownDescription: "Manage RustFS storage tiers for data transition to external backends",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Tier name (must be uppercase). Changing this forces recreation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"tier_type": schema.StringAttribute{
				Required:    true,
				Description: "Tier type: s3, minio, azure, gcs, aliyun, tencent, huaweicloud, r2, or rustfs.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"config_json": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Backend-specific configuration as a JSON string. See RustFS tier documentation for the expected format per backend type.",
			},
		},
	}
}

func (r *TierResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tierResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !json.Valid([]byte(plan.ConfigJson.ValueString())) {
		resp.Diagnostics.AddError("Invalid JSON", "config_json must be valid JSON")
		return
	}

	if err := r.client.RustClient.AddTier(json.RawMessage(plan.ConfigJson.ValueString())); err != nil {
		resp.Diagnostics.AddError("Error adding tier", "Could not add tier: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tierResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tierResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !json.Valid([]byte(plan.ConfigJson.ValueString())) {
		resp.Diagnostics.AddError("Invalid JSON", "config_json must be valid JSON")
		return
	}

	if err := r.client.RustClient.EditTier(plan.Name.ValueString(), json.RawMessage(plan.ConfigJson.ValueString())); err != nil {
		resp.Diagnostics.AddError("Error editing tier", "Could not edit tier: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data tierResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.RemoveTier(data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error removing tier", "Could not remove tier: "+err.Error())
		return
	}
}

func (r *TierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
