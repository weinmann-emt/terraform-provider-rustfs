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

var (
	_ resource.Resource                = &GroupPolicyAttachmentRessource{}
	_ resource.ResourceWithImportState = &GroupPolicyAttachmentRessource{}
)

type GroupPolicyAttachmentRessource struct {
	client *AllClient
}

type GroupPolicyAttachmentRessourceModel struct {
	Group  types.String `tfsdk:"group"`
	Policy types.String `tfsdk:"policy"`
	ID     types.String `tfsdk:"id"`
}

func NewGroupPolicyAttachmentRessource() resource.Resource {
	return &GroupPolicyAttachmentRessource{}
}

func (r *GroupPolicyAttachmentRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_policy_attachment"
}

func (r *GroupPolicyAttachmentRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Attach a canned IAM policy to an IAM group",
		MarkdownDescription: "Attach a canned IAM policy to an IAM group and detach it on destroy",
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IAM group. Changing this forces a new attachment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.StringAttribute{
				Required:    true,
				Description: "Name of the canned policy. Changing this forces a new attachment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form group/policy.",
			},
		},
	}
}

func (r *GroupPolicyAttachmentRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupPolicyAttachmentRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.AttachGroupPolicy(plan.Group.ValueString(), plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching policy to group",
			"Could not attach policy: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(plan.Group.ValueString() + "/" + plan.Policy.ValueString())
	tflog.Trace(ctx, "attached policy to group")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupPolicyAttachmentRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The API exposes no read-back for group policy attachments; the attach
	// operation is idempotent and the natural key is immutable, so the last
	// known state is preserved (same pattern as rustfs_bucket).
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupPolicyAttachmentRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// All attributes are RequiresReplace, so Terraform normally plans a
	// replacement; re-attaching keeps this method correct if it is invoked.
	err := r.client.RustClient.AttachGroupPolicy(plan.Group.ValueString(), plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching policy to group",
			"Could not attach policy: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(plan.Group.ValueString() + "/" + plan.Policy.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupPolicyAttachmentRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.DetachGroupPolicy(data.Group.ValueString(), data.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error detaching policy from group",
			"Could not detach policy: "+err.Error(),
		)
		return
	}
}

func (r *GroupPolicyAttachmentRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected format: <group>/<policy>",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
