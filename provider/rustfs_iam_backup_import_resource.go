package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IamBackupImportResource{}

type IamBackupImportResource struct {
	client *AllClient
}

type IamBackupImportResourceModel struct {
	ContentBase64 types.String `tfsdk:"content_base64"`
}

func NewIamBackupImportResource() resource.Resource {
	return &IamBackupImportResource{}
}

func (r *IamBackupImportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_backup_import"
}

func (r *IamBackupImportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Import RustFS IAM data from a ZIP archive",
		MarkdownDescription: "Import RustFS IAM entities (users, groups, policies, service accounts) from a base64-encoded ZIP archive",
		Attributes: map[string]schema.Attribute{
			"content_base64": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Base64-encoded ZIP archive containing IAM data to import. Changing this forces recreation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *IamBackupImportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IamBackupImportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IamBackupImportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := base64.StdEncoding.DecodeString(plan.ContentBase64.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid base64 data", err.Error())
		return
	}

	if err := r.client.RustClient.ImportIam(data); err != nil {
		resp.Diagnostics.AddError(
			"Error importing IAM data",
			"Could not import IAM data: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamBackupImportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IamBackupImportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IamBackupImportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IamBackupImportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := base64.StdEncoding.DecodeString(plan.ContentBase64.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid base64 data", err.Error())
		return
	}

	if err := r.client.RustClient.ImportIam(data); err != nil {
		resp.Diagnostics.AddError(
			"Error importing IAM data",
			"Could not import IAM data: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamBackupImportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
