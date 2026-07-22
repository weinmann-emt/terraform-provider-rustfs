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

var (
	_ resource.Resource                = &GroupResource{}
	_ resource.ResourceWithImportState = &GroupResource{}
)

type GroupResource struct {
	client *AllClient
}

type GroupResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Status  types.String `tfsdk:"status"`
	Members types.Set    `tfsdk:"members"`
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS IAM groups",
		MarkdownDescription: "Manage RustFS IAM groups and their members",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Group name. Changing this forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Group status: enabled or disabled. Defaults to enabled.",
			},
			"members": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Set of user access keys that are members of this group.",
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var members []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &members, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status := plan.Status.ValueString()
	if status == "" {
		status = "enabled"
	}

	if len(members) > 0 {
		err := r.client.RustClient.UpdateGroupMembers(rustfs.GroupAddRemove{
			Group:    plan.Name.ValueString(),
			Members:  members,
			IsRemove: false,
			Status:   status,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating group",
				"Could not create group: "+err.Error(),
			)
			return
		}
	}

	if status != "enabled" {
		if err := r.client.RustClient.SetGroupStatus(plan.Name.ValueString(), status); err != nil {
			resp.Diagnostics.AddError(
				"Error setting group status",
				"Could not set group status: "+err.Error(),
			)
			return
		}
	}

	plan.Status = types.StringValue(status)
	tflog.Trace(ctx, "created group resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := r.client.RustClient.GetGroup(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group",
			"Could not read group: "+err.Error(),
		)
		return
	}

	state.Status = types.StringValue(info.Status)
	members, diags := types.SetValueFrom(ctx, types.StringType, info.Members)
	resp.Diagnostics.Append(diags...)
	state.Members = members

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planMembers []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &planMembers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.DeleteGroup(plan.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error updating group",
			"Could not reset group members: "+err.Error(),
		)
		return
	}

	if len(planMembers) > 0 {
		status := plan.Status.ValueString()
		if status == "" {
			status = "enabled"
		}
		if err := r.client.RustClient.UpdateGroupMembers(rustfs.GroupAddRemove{
			Group:    plan.Name.ValueString(),
			Members:  planMembers,
			IsRemove: false,
			Status:   status,
		}); err != nil {
			resp.Diagnostics.AddError(
				"Error updating group members",
				"Could not add group members: "+err.Error(),
			)
			return
		}
	}

	if plan.Status.ValueString() != "" {
		if err := r.client.RustClient.SetGroupStatus(plan.Name.ValueString(), plan.Status.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error updating group status",
				"Could not set group status: "+err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.DeleteGroup(data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting group",
			"Could not delete group: "+err.Error(),
		)
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
