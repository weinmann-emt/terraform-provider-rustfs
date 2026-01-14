package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// Data models
type policyStatementModel struct {
	Effect    string   `tfsdk:"effect"`
	Action    []string `tfsdk:"action"`
	Ressource []string `tfsdk:"ressource"`
}

type policyResourceModel struct {
	Name      types.String           `tfsdk:"name"`
	Version   types.String           `tfsdk:"version"`
	Statement []policyStatementModel `tfsdk:"statement"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &PolicyRessource{}
)

// NewPolicyRessource is a helper function to simplify the provider implementation.
func NewPolicyRessource() resource.Resource {
	return &PolicyRessource{}
}

// PolicyRessource is the resource implementation.
type PolicyRessource struct {
	client *AllClient
}

// Metadata returns the resource type name.
func (r *PolicyRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// Schema defines the schema for the resource.
func (r *PolicyRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage S3 policies",
		MarkdownDescription: "Manage S3 policies",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the policy",
			},
			"version": schema.StringAttribute{
				Computed: true,
				Default:  stringdefault.StaticString("2012-10-17"),
			},
			"statement": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							Required: true,
						},
						"action": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"ressource": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *PolicyRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *PolicyRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fmt.Print(plan)

	statements := []rustfs.PolicyStatement{}
	for _, i := range plan.Statement {
		statements = append(statements,
			rustfs.PolicyStatement{
				Effect:    i.Effect,
				Action:    i.Action,
				Ressource: i.Ressource,
			},
		)
	}
	policy := rustfs.Policy{
		Version:   plan.Version.ValueString(),
		Name:      plan.Name.ValueString(),
		Statement: statements,
	}
	err := r.client.RustClient.CreatePolicy(policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating policy",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created a resource")
	// plan.ID = types.StringValue(account.AccessKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

// Read refreshes the Terraform state with the latest data.
func (r *PolicyRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyResourceModel
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
	actual, err := r.client.RustClient.ReadPolicy(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading policy",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(actual.Name)
	state.Version = types.StringValue(actual.Version)
	state.Statement = []policyStatementModel{}
	for _, read_statement := range actual.Statement {
		state.Statement = append(state.Statement,
			policyStatementModel{
				Effect:    read_statement.Effect,
				Action:    read_statement.Action,
				Ressource: read_statement.Ressource,
			},
		)
	}
	// Save update status
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *PolicyRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *PolicyRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data policyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.DeletePolicy(data.Name.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
	}
}
