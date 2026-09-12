package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &kmsKeyRessource{}
	_ resource.ResourceWithImportState = &kmsKeyRessource{}
)

// NewKmsKeyRessource is a helper function to simplify the provider implementation.
func NewKmsKeyRessource() resource.Resource {
	return &kmsKeyRessource{}
}

// kmsKeyRessource is the resource implementation.
type kmsKeyRessource struct {
	client *AllClient
}

type kmsKeyRessourceModel struct {
	Name        types.String `tfsdk:"name"`
	KeyID       types.String `tfsdk:"key_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	SkipDestroy types.Bool   `tfsdk:"skip_destroy"`
}

// Metadata returns the resource type name.
func (r *kmsKeyRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms_key"
}

// Schema defines the schema for the resource.
func (r *kmsKeyRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS KMS master keys",
		MarkdownDescription: "Manage RustFS KMS master keys. Deleting a KMS master key is **critical and irreversible**: it schedules the destruction of the key material and makes every object encrypted under the key permanently unreadable. Guard the key with `skip_destroy = true`.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "Name of the KMS master key. Changing this forces recreation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"key_id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-generated id of the KMS master key. On the Local dev backend this equals the key name.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the key was created.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the KMS master key is enabled. Defaults to `true`.",
			},
			"skip_destroy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When `true`, destroy only removes Terraform state and leaves the irreversible server-side key deletion unexecuted. Defaults to `false`.",
			},
		},
	}
}

func (r *kmsKeyRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the KMS key and sets the initial Terraform state.
func (r *kmsKeyRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kmsKeyRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.RustClient.CreateKmsKey(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating KMS key",
			"Could not create KMS key, unexpected error: "+err.Error(),
		)
		return
	}
	plan.KeyID = types.StringValue(key.KeyID)
	plan.CreatedAt = types.StringValue(key.CreationDate)

	if !plan.Enabled.ValueBool() {
		if err := r.client.RustClient.DisableKmsKey(key.KeyID); err != nil {
			resp.Diagnostics.AddError(
				"Error disabling KMS key",
				"Could not disable KMS key, unexpected error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, "created a KMS key")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *kmsKeyRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state kmsKeyRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.KeyID.ValueString()
	if keyID == "" {
		keyID = r.resolveKeyID(ctx, state.Name.ValueString(), resp)
		if resp.Diagnostics.HasError() {
			return
		}
		if keyID == "" {
			resp.State.RemoveResource(ctx)
			return
		}
		state.KeyID = types.StringValue(keyID)
	}

	key, err := r.client.RustClient.DescribeKmsKey(keyID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading KMS key",
			"Could not read KMS key, unexpected error: "+err.Error(),
		)
		return
	}
	state.Enabled = types.BoolValue(key.KeyState == "Enabled")
	state.CreatedAt = types.StringValue(key.CreationDate)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// resolveKeyID maps a key name to its server-side key id by listing the keys
// when the state does not yet carry one (import path). On the Local dev
// backend the key id equals the name; on production backends the key created
// through this provider carries a `name` tag that the listing exposes.
func (r *kmsKeyRessource) resolveKeyID(ctx context.Context, name string, resp *resource.ReadResponse) string {
	keys, err := r.client.RustClient.ListKmsKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing KMS keys",
			"Could not resolve key id from name, unexpected error: "+err.Error(),
		)
		return ""
	}
	for _, key := range keys {
		if key.KeyID == name || key.Tags["name"] == name {
			return key.KeyID
		}
	}
	tflog.Warn(ctx, "no KMS key matched name, removing from state", map[string]any{"name": name})
	return ""
}

// Update enables or disables the key and updates the Terraform state on success.
func (r *kmsKeyRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state kmsKeyRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Enabled.Equal(state.Enabled) {
		keyID := plan.KeyID.ValueString()
		if keyID == "" {
			keyID = plan.Name.ValueString()
		}
		var err error
		if plan.Enabled.ValueBool() {
			err = r.client.RustClient.EnableKmsKey(keyID)
		} else {
			err = r.client.RustClient.DisableKmsKey(keyID)
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating KMS key",
				"Could not update KMS key, unexpected error: "+err.Error(),
			)
			return
		}
	}
	// Computed-only attributes are carried over from the prior state: the
	// enable/disable switch does not change the key id or creation date.
	plan.KeyID = state.KeyID
	plan.CreatedAt = state.CreatedAt
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete irreversibly deletes the KMS key unless skip_destroy is set.
func (r *kmsKeyRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kmsKeyRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		tflog.Warn(ctx, "skip_destroy=true: leaving KMS key in place, removing only Terraform state")
		return
	}

	keyID := state.KeyID.ValueString()
	if keyID == "" {
		keyID = state.Name.ValueString()
	}
	if err := r.client.RustClient.DeleteKmsKey(keyID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting KMS key",
			"Could not delete KMS key, unexpected error: "+err.Error(),
		)
	}
}

func (r *kmsKeyRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
