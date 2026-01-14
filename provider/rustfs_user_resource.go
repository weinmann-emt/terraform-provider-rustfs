package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RustfsUserRessource{}
var _ resource.ResourceWithImportState = &RustfsUserRessource{}

// ExampleResource defines the resource implementation.
type RustfsUserRessource struct {
	client *AllClient
}

type RustfsUserRessourceModel struct {
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
	Status    types.String `tfsdk:"status"`
	Policy    types.String `tfsdk:"policy"`
	// ID        types.String `tfsdk:"id"`
	// Id        types.String `tfsdk:"id"`
}

func NewUserRessource() resource.Resource {
	return &RustfsUserRessource{}
}
func (r *RustfsUserRessource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *RustfsUserRessource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "RustFS user",

		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				MarkdownDescription: "Access Key",
				Required:            true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "Secret Key",
				Required:            true,
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status",
				Default:             stringdefault.StaticString("enabled"),
			},
			"policy": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "User policy",
			},
		},
	}
}
func (r *RustfsUserRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client

}
func (r *RustfsUserRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan RustfsUserRessourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account := rustfs.UserAccount{
		AccessKey: plan.AccessKey.ValueString(),
		SecretKey: plan.SecretKey.ValueString(),
		Policy:    plan.Policy.ValueString(),
	}

	err := r.client.RustClient.CreateUserAccount(account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created a resource")
	// plan.ID = types.StringValue(account.AccessKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	// plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *RustfsUserRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RustfsUserRessourceModel
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
	read, err := r.client.RustClient.ReadUserAccount(state.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading user",
			"Could read: "+err.Error(),
		)
		return
	}
	state.Status = types.StringValue(read.Status)
	state.AccessKey = types.StringValue(state.AccessKey.ValueString())
	state.SecretKey = types.StringValue(state.SecretKey.ValueString())
	state.Policy = types.StringValue(read.Policy)
	// state.ID = types.StringValue(state.ID.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
func (r *RustfsUserRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RustfsUserRessourceModel
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
	// plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}

func (r *RustfsUserRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RustfsUserRessourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	account := rustfs.UserAccount{
		AccessKey: data.AccessKey.ValueString(),
	}
	err := r.client.RustClient.DeleteUserAccount(account)
	if err != nil {
		tflog.Error(ctx, err.Error())
	}
}

func (r *RustfsUserRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
