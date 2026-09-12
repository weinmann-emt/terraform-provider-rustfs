package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var (
	_ resource.Resource                = &ModuleSwitchRessource{}
	_ resource.ResourceWithImportState = &ModuleSwitchRessource{}
)

type ModuleSwitchRessource struct {
	client *AllClient
}

type ModuleSwitchRessourceModel struct {
	ID            types.String `tfsdk:"id"`
	NotifyEnabled types.Bool   `tfsdk:"notify_enabled"`
	AuditEnabled  types.Bool   `tfsdk:"audit_enabled"`
}

func NewModuleSwitchRessource() resource.Resource {
	return &ModuleSwitchRessource{}
}

func (r *ModuleSwitchRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_switch"
}

func (r *ModuleSwitchRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS feature module switches",
		MarkdownDescription: "Manage RustFS feature module switches (notify and audit modules).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Fixed identifier for the module switch set.",
			},
			"notify_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the notify module is enabled.",
			},
			"audit_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the audit module is enabled.",
			},
		},
	}
}

func (r *ModuleSwitchRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *AllClient, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ModuleSwitchRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModuleSwitchRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.client.RustClient.SetModuleSwitches(moduleSwitchUpdateFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error setting module switches", "Could not set module switches: "+err.Error())
		return
	}

	tflog.Trace(ctx, "created module switches")
	plan.ID = types.StringValue("module-switches")
	applyModuleSwitchState(&plan, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ModuleSwitchRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModuleSwitchRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switches, err := r.client.RustClient.GetModuleSwitches()
	if err != nil {
		resp.Diagnostics.AddError("Error reading module switches", "Could not read module switches: "+err.Error())
		return
	}

	applyModuleSwitchState(&state, switches)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ModuleSwitchRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModuleSwitchRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.client.RustClient.SetModuleSwitches(moduleSwitchUpdateFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating module switches", "Could not update module switches: "+err.Error())
		return
	}

	plan.ID = types.StringValue("module-switches")
	applyModuleSwitchState(&plan, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the resource from state. The admin API exposes no reset
// endpoint for module switches, so the server switches are intentionally left
// unchanged on destroy.
func (r *ModuleSwitchRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "module switches left unchanged on destroy (no DELETE endpoint)")
}

func (r *ModuleSwitchRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func moduleSwitchUpdateFromModel(model ModuleSwitchRessourceModel) rustfs.ModuleSwitchUpdate {
	return rustfs.ModuleSwitchUpdate{
		NotifyEnabled: model.NotifyEnabled.ValueBool(),
		AuditEnabled:  model.AuditEnabled.ValueBool(),
	}
}

func applyModuleSwitchState(model *ModuleSwitchRessourceModel, state *rustfs.ModuleSwitchState) {
	model.NotifyEnabled = types.BoolValue(state.NotifyEnabled)
	model.AuditEnabled = types.BoolValue(state.AuditEnabled)
}
