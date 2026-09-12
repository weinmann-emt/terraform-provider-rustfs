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
	_ resource.Resource                = &UserPolicyAttachmentRessource{}
	_ resource.ResourceWithImportState = &UserPolicyAttachmentRessource{}
)

// UserPolicyAttachmentRessource attaches one canned IAM policy to a user on
// top of the user's primary policy, so a user can hold more than one policy.
type UserPolicyAttachmentRessource struct {
	client *AllClient
}

type UserPolicyAttachmentRessourceModel struct {
	User   types.String `tfsdk:"user"`
	Policy types.String `tfsdk:"policy"`
	ID     types.String `tfsdk:"id"`
}

// NewUserPolicyAttachmentRessource is a helper function to simplify the provider implementation.
func NewUserPolicyAttachmentRessource() resource.Resource {
	return &UserPolicyAttachmentRessource{}
}

// Metadata returns the resource type name.
func (r *UserPolicyAttachmentRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_policy_attachment"
}

// Schema defines the schema for the resource.
func (r *UserPolicyAttachmentRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Attach a canned IAM policy to a user",
		MarkdownDescription: "Attach a canned IAM policy to a user and detach it on destroy",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Required:    true,
				Description: "Access key of the IAM user. Changing this forces a new attachment.",
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
				Description: "Composite identifier in the form user/policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UserPolicyAttachmentRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create attaches the policy to the user and sets the initial Terraform state.
func (r *UserPolicyAttachmentRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.AttachUserPolicy(plan.User.ValueString(), plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching policy to user",
			"Could not attach policy: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(plan.User.ValueString() + "/" + plan.Policy.ValueString())
	tflog.Trace(ctx, "attached policy to user")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read preserves the last known state. The RustFS admin API exposes no
// read-back enumerating a user's extra policy attachments; the attach
// operation is idempotent and the user/policy pair is immutable, so the
// last known state is kept (same pattern as rustfs_bucket).
func (r *UserPolicyAttachmentRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = types.StringValue(state.User.ValueString() + "/" + state.Policy.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update re-attaches the policy. All attributes are RequiresReplace, so
// Terraform normally plans a replacement; re-attaching keeps this method
// correct if it is invoked.
func (r *UserPolicyAttachmentRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.AttachUserPolicy(plan.User.ValueString(), plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching policy to user",
			"Could not attach policy: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(plan.User.ValueString() + "/" + plan.Policy.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete detaches the policy from the user and removes the Terraform state on success.
func (r *UserPolicyAttachmentRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.DetachUserPolicy(data.User.ValueString(), data.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error detaching policy from user",
			"Could not detach policy: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "detached policy from user")
}

// ImportState supports importing an attachment using the composite
// <user>/<policy> identifier.
func (r *UserPolicyAttachmentRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected format: <user>/<policy>",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy"), parts[1])...)
}
