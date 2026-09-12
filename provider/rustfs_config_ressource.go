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
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var (
	_ resource.Resource                = &ConfigRessource{}
	_ resource.ResourceWithImportState = &ConfigRessource{}
)

type ConfigRessourceModel struct {
	SubSystem types.String `tfsdk:"sub_system"`
	Settings  types.Map    `tfsdk:"settings"`
	ID        types.String `tfsdk:"id"`
}

func NewConfigRessource() resource.Resource {
	return &ConfigRessource{}
}

type ConfigRessource struct {
	client *AllClient
}

func (r *ConfigRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (r *ConfigRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage server sub-system configuration (config-kv)",
		MarkdownDescription: "Manage server sub-system configuration (config-kv), the `mc admin config` equivalent. Each resource manages one sub-system scope.",
		Attributes: map[string]schema.Attribute{
			"sub_system": schema.StringAttribute{
				Required:    true,
				Description: "Sub-system scope to manage, e.g. `notify_webhook` or `notify_webhook:primary`. Changing this forces recreation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"settings": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Key/value settings applied to the sub-system scope.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (the sub-system scope).",
			},
		},
	}
}

func (r *ConfigRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func configKVsFromSettings(settings types.Map) ([]rustfs.ConfigKV, error) {
	elements := settings.Elements()
	kvs := make([]rustfs.ConfigKV, 0, len(elements))
	for key, value := range elements {
		str, ok := value.(types.String)
		if !ok {
			return nil, fmt.Errorf("unexpected value type for setting %q: %T", key, value)
		}
		kvs = append(kvs, rustfs.ConfigKV{Key: key, Value: str.ValueString()})
	}
	return kvs, nil
}

func settingsFromConfigKVs(ctx context.Context, kvs []rustfs.ConfigKV) types.Map {
	settings := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		settings[kv.Key] = kv.Value
	}
	value, diags := types.MapValueFrom(ctx, types.StringType, settings)
	if diags.HasError() {
		return types.MapNull(types.StringType)
	}
	return value
}

func (r *ConfigRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConfigRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kvs, err := configKVsFromSettings(plan.Settings)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing settings", err.Error())
		return
	}
	if err := r.client.RustClient.SetConfig(plan.SubSystem.ValueString(), kvs); err != nil {
		resp.Diagnostics.AddError(
			"Error setting config",
			"Could not set config, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = plan.SubSystem
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConfigRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kvs, err := r.client.RustClient.GetConfig(state.SubSystem.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading config",
			"Could not read config, unexpected error: "+err.Error(),
		)
		return
	}

	state.Settings = settingsFromConfigKVs(ctx, kvs)
	state.ID = state.SubSystem
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConfigRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConfigRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kvs, err := configKVsFromSettings(plan.Settings)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing settings", err.Error())
		return
	}
	if err := r.client.RustClient.SetConfig(plan.SubSystem.ValueString(), kvs); err != nil {
		resp.Diagnostics.AddError(
			"Error updating config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = plan.SubSystem
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.DeleteConfig(data.SubSystem.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
	}
}

func (r *ConfigRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("sub_system"), req, resp)
}
