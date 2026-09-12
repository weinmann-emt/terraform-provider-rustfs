package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var (
	_ resource.Resource                = &LDAPPolicyAttachmentRessource{}
	_ resource.ResourceWithImportState = &LDAPPolicyAttachmentRessource{}
)

type LDAPPolicyAttachmentRessource struct {
	client *AllClient
}

type LDAPPolicyAttachmentRessourceModel struct {
	UserOrGroup types.String `tfsdk:"user_or_group"`
	Policy      types.String `tfsdk:"policy"`
	IsGroup     types.Bool   `tfsdk:"is_group"`
	ID          types.String `tfsdk:"id"`
}

func NewLDAPPolicyAttachmentRessource() resource.Resource {
	return &LDAPPolicyAttachmentRessource{}
}

func (r *LDAPPolicyAttachmentRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_policy_attachment"
}

func (r *LDAPPolicyAttachmentRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Attach a canned policy to an LDAP user or group",
		MarkdownDescription: "Attach a canned policy to an LDAP user or group and detach it on destroy",
		Attributes: map[string]schema.Attribute{
			"user_or_group": schema.StringAttribute{
				Required:    true,
				Description: "LDAP user or group (distinguished name) the policy is attached to. Changing this forces a new attachment.",
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
			"is_group": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether user_or_group is an LDAP group. Defaults to false (LDAP user).",
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form user_or_group/policy.",
			},
		},
	}
}

func (r *LDAPPolicyAttachmentRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LDAPPolicyAttachmentRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LDAPPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.AttachLDAPPolicy(rustfs.LDAPPolicyAttachment{
		UserOrGroup: plan.UserOrGroup.ValueString(),
		PolicyName:  plan.Policy.ValueString(),
		IsGroup:     plan.IsGroup.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching policy to LDAP entity",
			"Could not attach policy: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(plan.UserOrGroup.ValueString() + "/" + plan.Policy.ValueString())
	tflog.Trace(ctx, "attached policy to LDAP entity")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LDAPPolicyAttachmentRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LDAPPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// There is no per-target read-back endpoint for LDAP policy attachments
	// and the attach operation is idempotent with an immutable natural key,
	// so the last known state is preserved (same pattern as rustfs_bucket).
	// The identifier and default are recomputed so imported state round-trips.
	if state.IsGroup.IsNull() {
		state.IsGroup = types.BoolValue(false)
	}
	state.ID = types.StringValue(state.UserOrGroup.ValueString() + "/" + state.Policy.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LDAPPolicyAttachmentRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LDAPPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// All attributes are RequiresReplace, so Terraform normally plans a
	// replacement; re-attaching keeps this method correct if it is invoked.
	err := r.client.RustClient.AttachLDAPPolicy(rustfs.LDAPPolicyAttachment{
		UserOrGroup: plan.UserOrGroup.ValueString(),
		PolicyName:  plan.Policy.ValueString(),
		IsGroup:     plan.IsGroup.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching policy to LDAP entity",
			"Could not attach policy: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(plan.UserOrGroup.ValueString() + "/" + plan.Policy.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LDAPPolicyAttachmentRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LDAPPolicyAttachmentRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.DetachLDAPPolicy(rustfs.LDAPPolicyAttachment{
		UserOrGroup: data.UserOrGroup.ValueString(),
		PolicyName:  data.Policy.ValueString(),
		IsGroup:     data.IsGroup.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error detaching policy from LDAP entity",
			"Could not detach policy: "+err.Error(),
		)
		return
	}
}

func (r *LDAPPolicyAttachmentRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Accepts "<user_or_group>/<policy>" and optionally
	// "<user_or_group>/<policy>/<is_group>". The policy is identified from the
	// right so distinguished names containing "/" are still supported.
	parts := strings.Split(req.ID, "/")
	if len(parts) < 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected format: <user_or_group>/<policy> or <user_or_group>/<policy>/<is_group>",
		)
		return
	}
	isGroupIdx := -1
	policyIdx := len(parts) - 1
	if len(parts) >= 3 && (parts[len(parts)-2] == "true" || parts[len(parts)-2] == "false") {
		isGroupIdx = len(parts) - 2
		policyIdx = len(parts) - 2
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_or_group"), strings.Join(parts[:policyIdx], "/"))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy"), parts[policyIdx])...)
	if isGroupIdx != -1 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("is_group"), parts[isGroupIdx])...)
	}
}
