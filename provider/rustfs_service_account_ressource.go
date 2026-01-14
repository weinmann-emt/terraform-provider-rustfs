// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

type serviceAccountResourceModel struct {
	AccessKey   types.String `tfsdk:"access_key"`
	SecretKey   types.String `tfsdk:"secret_key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	TargetUser  types.String `tfsdk:"user"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &ServiceAccountRessource{}
)

// NewServiceAccountRessource is a helper function to simplify the provider implementation.
func NewServiceAccountRessource() resource.Resource {
	return &ServiceAccountRessource{}
}

// ServiceAccountRessource is the resource implementation.
type ServiceAccountRessource struct {
	client *AllClient
}

// Metadata returns the resource type name.
func (r *ServiceAccountRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceaccount"
}

// Schema defines the schema for the resource.
func (r *ServiceAccountRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage ServiceUser/API Keys",
		MarkdownDescription: "Manage ServiceUser/API Keys",
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				MarkdownDescription: "Access Key",
				Required:            true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "Secret Key",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Visible name, only for viewing",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Short description of the scope we plan to use this token",
			},
			"user": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional user the token should be scoped to",
			},
		},
	}
}

func (r *ServiceAccountRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AllClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *ServiceAccountRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan serviceAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account := rustfs.ServiceAccount{
		Name:        plan.Name.ValueString(),
		AccessKey:   plan.AccessKey.ValueString(),
		SecretKey:   plan.SecretKey.ValueString(),
		Description: plan.Description.ValueString(),
		TargetUser:  plan.TargetUser.ValueString(),
	}
	err := r.client.RustClient.CreateServiceAccount(account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating policy",
			"Could not create service account, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

// Read refreshes the Terraform state with the latest data.
func (r *ServiceAccountRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make the read request
	actual, err := r.client.RustClient.ReadServiceAccount(state.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading policy",
			"Could not read service avvount, unexpected error: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(actual.Name)
	state.Description = types.StringValue(actual.Description)
	// Save update status
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ServiceAccountRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account := rustfs.ServiceAccount{
		Name:        plan.Name.ValueString(),
		AccessKey:   plan.AccessKey.ValueString(),
		SecretKey:   plan.SecretKey.ValueString(),
		Description: plan.Description.ValueString(),
	}
	err := r.client.RustClient.UpdateServiceAccount(account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating policy",
			"Could not update order, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(account.Name)
	plan.Description = types.StringValue(account.Description)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ServiceAccountRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serviceAccountResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	account := rustfs.ServiceAccount{
		Name:        data.Name.ValueString(),
		AccessKey:   data.AccessKey.ValueString(),
		SecretKey:   data.SecretKey.ValueString(),
		Description: data.Description.ValueString(),
	}
	err := r.client.RustClient.DeleteServiceAccount(account)
	if err != nil {
		tflog.Error(ctx, err.Error())
	}
}
