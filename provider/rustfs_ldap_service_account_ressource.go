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
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &LDAPServiceAccountRessource{}
	_ resource.ResourceWithImportState = &LDAPServiceAccountRessource{}
)

// NewLDAPServiceAccountRessource is a helper function to simplify the provider implementation.
func NewLDAPServiceAccountRessource() resource.Resource {
	return &LDAPServiceAccountRessource{}
}

// LDAPServiceAccountRessource is the resource implementation.
type LDAPServiceAccountRessource struct {
	client *AllClient
}

type LDAPServiceAccountRessourceModel struct {
	AccessKey   types.String `tfsdk:"access_key"`
	SecretKey   types.String `tfsdk:"secret_key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	User        types.String `tfsdk:"user"`
	Policy      types.String `tfsdk:"policy"`
}

// Metadata returns the resource type name.
func (r *LDAPServiceAccountRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_service_account"
}

// Schema defines the schema for the resource.
func (r *LDAPServiceAccountRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage service accounts scoped to LDAP users",
		MarkdownDescription: "Create a service account scoped to an LDAP user",
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Access key of the service account. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secret_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Secret key of the service account.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Visible name, only for viewing.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Short description of the scope we plan to use this token.",
			},
			"user": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "LDAP user (distinguished name) the service account is scoped to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IAM policy document (JSON) the service account is scoped to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *LDAPServiceAccountRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *LDAPServiceAccountRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LDAPServiceAccountRessourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account := rustfs.ServiceAccount{
		AccessKey:   plan.AccessKey.ValueString(),
		SecretKey:   plan.SecretKey.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		TargetUser:  plan.User.ValueString(),
		Policy:      plan.Policy.ValueString(),
	}
	err := r.client.RustClient.CreateLDAPServiceAccount(account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating LDAP service account",
			"Could not create service account, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created an LDAP service account")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *LDAPServiceAccountRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LDAPServiceAccountRessourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The LDAP endpoint family exposes no dedicated read-back, so the shared
	// info-service-account endpoint is used. An LDAP service account is a
	// standard service account, and parentUser carries the LDAP user DN it is
	// scoped to.
	actual, err := r.client.RustClient.ReadServiceAccount(state.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LDAP service account",
			"Could not read service account, unexpected error: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(actual.Name)
	if actual.Description != "" {
		state.Description = types.StringValue(actual.Description)
	}
	if actual.ParentUser != "" {
		state.User = types.StringValue(actual.ParentUser)
	}
	if !actual.ImpliedPolicy && state.Policy.IsNull() {
		state.Policy = types.StringValue(actual.Policy)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *LDAPServiceAccountRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LDAPServiceAccountRessourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account := rustfs.ServiceAccount{
		AccessKey:   plan.AccessKey.ValueString(),
		SecretKey:   plan.SecretKey.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	err := r.client.RustClient.UpdateServiceAccount(account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating LDAP service account",
			"Could not update service account, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *LDAPServiceAccountRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LDAPServiceAccountRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No LDAP-specific delete endpoint exists; a created LDAP service account
	// is a standard service account and is removed via delete-service-accounts.
	account := rustfs.ServiceAccount{
		AccessKey: data.AccessKey.ValueString(),
	}
	err := r.client.RustClient.DeleteServiceAccount(account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting LDAP service account",
			"Could not delete service account, unexpected error: "+err.Error(),
		)
	}
}

func (r *LDAPServiceAccountRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("access_key"), req, resp)
}
