package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

type serviceAccountPolicyStatementModel struct {
	Effect   types.String `tfsdk:"effect"`
	Action   types.Set    `tfsdk:"action"`
	Resource types.Set    `tfsdk:"resource"`
}

type serviceAccountPolicyModel struct {
	Version   types.String `tfsdk:"version"`
	Statement types.List   `tfsdk:"statement"`
}

type serviceAccountResourceModel struct {
	AccessKey   types.String `tfsdk:"access_key"`
	SecretKey   types.String `tfsdk:"secret_key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	TargetUser  types.String `tfsdk:"user"`
	Policy      types.Object `tfsdk:"policy"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ServiceAccountRessource{}
	_ resource.ResourceWithImportState = &ServiceAccountRessource{}
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
				Computed:            true,
				MarkdownDescription: "Optional user the token should be scoped to. Changing this forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Inline policy scoping this service account's permissions. If omitted, the account inherits the target user's policy.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Computed: true,
						Default:  stringdefault.StaticString("2012-10-17"),
					},
					"statement": schema.ListNestedAttribute{
						Required:            true,
						MarkdownDescription: "The single statement that makes up this policy.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"effect": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf("Allow", "Deny"),
									},
								},
								"action": schema.SetAttribute{
									ElementType: types.StringType,
									Required:    true,
								},
								"resource": schema.SetAttribute{
									ElementType: types.StringType,
									Required:    true,
								},
							},
						},
					},
				},
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
			"Unexpected Resource Configure Type",
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

	account := rustfs.ServiceAccountCreate{
		Name:        plan.Name.ValueString(),
		AccessKey:   plan.AccessKey.ValueString(),
		SecretKey:   plan.SecretKey.ValueString(),
		Description: plan.Description.ValueString(),
		TargetUser:  plan.TargetUser.ValueString(),
	}

	if !plan.Policy.IsNull() && !plan.Policy.IsUnknown() {
		var policyObj serviceAccountPolicyModel
		diags := plan.Policy.As(ctx, &policyObj, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		jsonPolicy, policyDiags := convertPolicyToJSON(ctx, &policyObj)
		resp.Diagnostics.Append(policyDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		account.Policy = jsonPolicy
	}

	err := r.client.RustClient.CreateServiceAccount(account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service account",
			"Could not create service account, unexpected error: "+err.Error(),
		)
		return
	}

	infoResp, err := r.client.RustClient.ReadServiceAccount(account.AccessKey)
	if err != nil {
		resp.Diagnostics.AddError("API Read-After-Create Error", err.Error())
		return
	}

	policyModel, policyDiags := parsePolicyJSON(ctx, infoResp.Policy)
	resp.Diagnostics.Append(policyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectVal, objectDiags := types.ObjectValueFrom(ctx, plan.Policy.AttributeTypes(ctx), policyModel)
	resp.Diagnostics.Append(objectDiags...)
	plan.Policy = objectVal

	plan.TargetUser = types.StringValue(infoResp.ParentUser)

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
			"Error reading service account",
			"Could not read service account, unexpected error: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(actual.Name)
	state.Description = types.StringValue(actual.Description)
	if actual.ParentUser == "" {
		state.TargetUser = types.StringNull()
	} else {
		state.TargetUser = types.StringValue(actual.ParentUser)
	}

	policyModel, policyDiags := parsePolicyJSON(ctx, actual.Policy)
	resp.Diagnostics.Append(policyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectVal, objectDiags := types.ObjectValueFrom(ctx, state.Policy.AttributeTypes(ctx), policyModel)
	resp.Diagnostics.Append(objectDiags...)
	state.Policy = objectVal

	// Save update status
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

	account := rustfs.ServiceAccountUpdate{
		NewName:        plan.Name.ValueString(),
		NewDescription: plan.Description.ValueString(),
		NewSecretKey:   plan.SecretKey.ValueString(),
	}

	if !plan.Policy.IsNull() && !plan.Policy.IsUnknown() {
		var policyObj serviceAccountPolicyModel
		diags := plan.Policy.As(ctx, &policyObj, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		jsonPolicy, policyDiags := convertPolicyToJSON(ctx, &policyObj)
		resp.Diagnostics.Append(policyDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		account.NewPolicy = jsonPolicy
	}

	err := r.client.RustClient.UpdateServiceAccount(plan.AccessKey.ValueString(), account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service account",
			"Could not update service account, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(account.NewName)
	plan.Description = types.StringValue(account.NewDescription)

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

	err := r.client.RustClient.DeleteServiceAccount(data.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting service account",
			"Could not delete service account, unexpected error: "+err.Error(),
		)
	}
}

func (r *ServiceAccountRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("access_key"), req, resp)
}

func convertPolicyToJSON(ctx context.Context, policy *serviceAccountPolicyModel) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if policy == nil || policy.Version.IsNull() || policy.Version.IsUnknown() {
		return "", diags
	}

	type statementJSON struct {
		Effect   string   `json:"Effect"`
		Action   []string `json:"Action"`
		Resource []string `json:"Resource"`
	}

	type policyJSON struct {
		Version   string          `json:"Version"`
		Statement []statementJSON `json:"Statement"`
	}

	var statementModels []serviceAccountPolicyStatementModel
	if listDiags := policy.Statement.ElementsAs(ctx, &statementModels, false); listDiags.HasError() {
		diags.Append(listDiags...)
		return "", diags
	}

	var jsonStatements []statementJSON
	for _, s := range statementModels {
		var actions []string
		if actionDiags := s.Action.ElementsAs(ctx, &actions, false); actionDiags.HasError() {
			diags.Append(actionDiags...)
			return "", diags
		}

		var resources []string
		if resourceDiags := s.Resource.ElementsAs(ctx, &resources, false); resourceDiags.HasError() {
			diags.Append(resourceDiags...)
			return "", diags
		}

		jsonStatements = append(jsonStatements, statementJSON{
			Effect:   s.Effect.ValueString(),
			Action:   actions,
			Resource: resources,
		})
	}

	payload := policyJSON{
		Version:   policy.Version.ValueString(),
		Statement: jsonStatements,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		diags.AddError(
			"Error generating policy JSON",
			"Could not marshal policy structure into a valid JSON string: "+err.Error(),
		)
		return "", diags
	}

	return string(jsonBytes), diags
}

func parsePolicyJSON(ctx context.Context, rawPolicyStr string) (*serviceAccountPolicyModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if rawPolicyStr == "" {
		return nil, diags
	}

	var raw struct {
		Version   string `json:"Version"`
		Statement []struct {
			Effect   string   `json:"Effect"`
			Action   []string `json:"Action"`
			Resource []string `json:"Resource"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal([]byte(rawPolicyStr), &raw); err != nil {
		diags.AddError(
			"Error parsing policy JSON from API",
			"Could not unmarshal API policy string: "+err.Error(),
		)
		return nil, diags
	}

	if len(raw.Statement) == 0 {
		return nil, diags
	}

	var statementModels []serviceAccountPolicyStatementModel
	for _, s := range raw.Statement {
		actionSet, actionDiags := types.SetValueFrom(ctx, types.StringType, s.Action)
		diags.Append(actionDiags...)

		resourceSet, resourceDiags := types.SetValueFrom(ctx, types.StringType, s.Resource)
		diags.Append(resourceDiags...)

		statementModels = append(statementModels, serviceAccountPolicyStatementModel{
			Effect:   types.StringValue(s.Effect),
			Action:   actionSet,
			Resource: resourceSet,
		})
	}

	statementList, listDiags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"effect":   types.StringType,
			"action":   types.SetType{ElemType: types.StringType},
			"resource": types.SetType{ElemType: types.StringType},
		},
	}, statementModels)
	diags.Append(listDiags...)

	if diags.HasError() {
		return nil, diags
	}

	return &serviceAccountPolicyModel{
		Version:   types.StringValue(raw.Version),
		Statement: statementList,
	}, diags
}
