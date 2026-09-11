package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &auditTargetRessource{}
	_ resource.ResourceWithImportState = &auditTargetRessource{}
)

// NewAuditTargetRessource is a helper function to simplify the provider implementation.
func NewAuditTargetRessource() resource.Resource {
	return &auditTargetRessource{}
}

// auditTargetRessource is the resource implementation.
type auditTargetRessource struct {
	client *AllClient
}

type auditTargetRessourceModel struct {
	TargetType    types.String `tfsdk:"target_type"`
	TargetName    types.String `tfsdk:"target_name"`
	Endpoint      types.String `tfsdk:"endpoint"`
	AuthToken     types.String `tfsdk:"auth_token"`
	Comment       types.String `tfsdk:"comment"`
	QueueLimit    types.Int64  `tfsdk:"queue_limit"`
	QueueDir      types.String `tfsdk:"queue_dir"`
	ClientCert    types.String `tfsdk:"client_cert"`
	ClientKey     types.String `tfsdk:"client_key"`
	ClientCA      types.String `tfsdk:"client_ca"`
	SkipTLSVerify types.Bool   `tfsdk:"skip_tls_verify"`
	HealthState   types.String `tfsdk:"health_state"`
	HealthReason  types.String `tfsdk:"health_reason"`
	Status        types.String `tfsdk:"status"`
}

// Metadata returns the resource type name.
func (r *auditTargetRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_target"
}

// Schema defines the schema for the resource.
func (r *auditTargetRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS audit-log webhook/HTTP targets",
		MarkdownDescription: "Manage RustFS audit-log webhook/HTTP targets",
		Attributes: map[string]schema.Attribute{
			"target_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("audit_webhook"),
				Description: "Audit target subsystem. Defaults to audit_webhook.",
				Validators: []validator.String{stringvalidator.OneOf(
					"audit_webhook",
					"audit_kafka",
					"audit_amqp",
					"audit_mqtt",
					"audit_mysql",
					"audit_nats",
					"audit_postgres",
					"audit_pulsar",
					"audit_redis",
				)},
			},
			"target_name": schema.StringAttribute{
				Required:      true,
				Description:   "Name of the audit target. Changing this forces recreation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Endpoint URL the audit target posts to (webhook).",
			},
			"auth_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Optional bearer token for the audit target (webhook). Not refreshed from the server on read.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "Optional comment describing the audit target.",
			},
			"queue_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Optional in-memory queue limit for the audit target.",
			},
			"queue_dir": schema.StringAttribute{
				Optional:    true,
				Description: "Optional absolute path of the on-disk queue directory for the audit target.",
			},
			"client_cert": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Optional client certificate for mTLS to the audit target.",
			},
			"client_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Optional client private key for mTLS to the audit target.",
			},
			"client_ca": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Optional CA certificate used to verify the audit target.",
			},
			"skip_tls_verify": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Skip TLS certificate verification for the audit target.",
			},
			"health_state": schema.StringAttribute{
				Computed:    true,
				Description: "Runtime health state of the audit target (offline/online/...).",
			},
			"health_reason": schema.StringAttribute{
				Computed:    true,
				Description: "Runtime health reason for the audit target.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Runtime status of the audit target.",
			},
		},
	}
}

func (r *auditTargetRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// auditTargetKeyValues builds the key/value config body from the resource model,
// including only the fields the user has actually set.
func (r *auditTargetRessource) auditTargetKeyValues(plan *auditTargetRessourceModel) []rustfs.AuditTargetKeyValue {
	var kvs []rustfs.AuditTargetKeyValue
	add := func(key string, value string) {
		if value != "" {
			kvs = append(kvs, rustfs.AuditTargetKeyValue{Key: key, Value: value})
		}
	}
	add("endpoint", plan.Endpoint.ValueString())
	add("auth_token", plan.AuthToken.ValueString())
	add("comment", plan.Comment.ValueString())
	add("queue_dir", plan.QueueDir.ValueString())
	add("client_cert", plan.ClientCert.ValueString())
	add("client_key", plan.ClientKey.ValueString())
	add("client_ca", plan.ClientCA.ValueString())
	if !plan.QueueLimit.IsNull() && !plan.QueueLimit.IsUnknown() {
		kvs = append(kvs, rustfs.AuditTargetKeyValue{Key: "queue_limit", Value: strconv.FormatInt(plan.QueueLimit.ValueInt64(), 10)})
	}
	if !plan.SkipTLSVerify.IsNull() && !plan.SkipTLSVerify.IsUnknown() {
		kvs = append(kvs, rustfs.AuditTargetKeyValue{Key: "skip_tls_verify", Value: strconv.FormatBool(plan.SkipTLSVerify.ValueBool())})
	}
	return kvs
}

// serviceForType maps an audit target_type to the service reported by the list
// endpoint (audit_webhook -> webhook, audit_kafka -> kafka, ...).
func serviceForType(targetType string) string {
	return strings.TrimPrefix(targetType, "audit_")
}

// Create adds the audit target and sets the initial Terraform state.
func (r *auditTargetRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan auditTargetRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.SetAuditTarget(
		plan.TargetType.ValueString(),
		plan.TargetName.ValueString(),
		r.auditTargetKeyValues(&plan),
	); err != nil {
		resp.Diagnostics.AddError(
			"Error creating audit target",
			"Could not create audit target, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created an audit target")
	if err := r.refreshHealth(&plan); err != nil {
		resp.Diagnostics.AddError(
			"Error reading audit target",
			"Could not read audit target after create, unexpected error: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// refreshHealth lists the audit targets and populates the computed health
// metadata on the model when the target is present.
func (r *auditTargetRessource) refreshHealth(model *auditTargetRessourceModel) error {
	targets, err := r.client.RustClient.ListAuditTargets()
	if err != nil {
		return err
	}

	key := model.TargetName.ValueString()
	service := serviceForType(model.TargetType.ValueString())
	for i := range targets {
		t := &targets[i]
		if t.AccountID == key && (service == "" || t.Service == service) {
			model.Status = types.StringValue(t.Status)
			model.HealthState = types.StringValue(t.HealthState)
			model.HealthReason = types.StringValue(t.HealthReason)
			return nil
		}
	}
	return fmt.Errorf("audit target %s not found in list", key)
}

// Read refreshes the Terraform state with the latest data. The admin list
// endpoint does not echo the target configuration (endpoint, auth token, ...),
// only identity and health metadata, so the config fields are preserved from
// the prior state and only the health/computed metadata is refreshed.
func (r *auditTargetRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state auditTargetRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targets, err := r.client.RustClient.ListAuditTargets()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing audit targets",
			"Could not list audit targets, unexpected error: "+err.Error(),
		)
		return
	}

	key := state.TargetName.ValueString()
	service := serviceForType(state.TargetType.ValueString())
	var found *rustfs.AuditTarget
	for i := range targets {
		t := &targets[i]
		if t.AccountID == key && (service == "" || t.Service == service) {
			found = t
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Status = types.StringValue(found.Status)
	state.HealthState = types.StringValue(found.HealthState)
	state.HealthReason = types.StringValue(found.HealthReason)
	// endpoint, auth_token, comment and the other config fields are preserved
	// from state; the list response is not trusted to echo them back.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update edits the audit target and updates the Terraform state on success.
func (r *auditTargetRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan auditTargetRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.SetAuditTarget(
		plan.TargetType.ValueString(),
		plan.TargetName.ValueString(),
		r.auditTargetKeyValues(&plan),
	); err != nil {
		resp.Diagnostics.AddError(
			"Error updating audit target",
			"Could not update audit target, unexpected error: "+err.Error(),
		)
		return
	}
	if err := r.refreshHealth(&plan); err != nil {
		resp.Diagnostics.AddError(
			"Error reading audit target",
			"Could not read audit target after update, unexpected error: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resets the audit target to its default configuration and removes the
// Terraform state. The admin API exposes no plain delete for targets, so
// destroy goes through the reset endpoint.
func (r *auditTargetRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state auditTargetRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.ResetAuditTarget(
		state.TargetType.ValueString(),
		state.TargetName.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError(
			"Error resetting audit target",
			"Could not reset audit target, unexpected error: "+err.Error(),
		)
	}
}

func (r *auditTargetRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected the import ID to be <target_type>/<target_name>",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_name"), parts[1])...)
}
