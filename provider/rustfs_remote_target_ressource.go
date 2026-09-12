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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var (
	_ resource.Resource                = &RemoteTargetRessource{}
	_ resource.ResourceWithImportState = &RemoteTargetRessource{}
)

// RemoteTargetRessource manages remote targets used as replication destinations
// or notification ARNs in rustfs.
type RemoteTargetRessource struct {
	client *AllClient
}

type RemoteTargetRessourceModel struct {
	Arn          types.String `tfsdk:"arn"`
	Type         types.String `tfsdk:"type"`
	Endpoint     types.String `tfsdk:"endpoint"`
	AccessKey    types.String `tfsdk:"access_key"`
	SecretKey    types.String `tfsdk:"secret_key"`
	Secure       types.Bool   `tfsdk:"secure"`
	Region       types.String `tfsdk:"region"`
	Path         types.String `tfsdk:"path"`
	Bucket       types.String `tfsdk:"bucket"`
	TargetBucket types.String `tfsdk:"target_bucket"`
}

// NewRemoteTargetRessource returns a new RemoteTargetRessource.
func NewRemoteTargetRessource() resource.Resource {
	return &RemoteTargetRessource{}
}

func (r *RemoteTargetRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_target"
}

func (r *RemoteTargetRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage remote targets used as replication destinations or notification ARNs in rustfs",
		MarkdownDescription: "Manage remote targets used as replication destinations or notification ARNs in rustfs",
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed:    true,
				Description: "ARN assigned by the server to the remote target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("replication"),
				Description: "Remote target type. Only 'replication' is supported by this server version.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Endpoint of the remote target, reachable from the rustfs server.",
			},
			"access_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Access key for the remote target.",
			},
			"secret_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Secret key for the remote target.",
			},
			"secure": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Use TLS for the remote target connection. Defaults to true.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region of the remote target.",
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Description: "Path prefix on the remote target.",
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Source bucket on which the remote target is registered.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_bucket": schema.StringAttribute{
				Required:    true,
				Description: "Destination bucket on the remote target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RemoteTargetRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RemoteTargetRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RemoteTargetRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	arn, err := r.client.RustClient.AddRemoteTarget(plan.Bucket.ValueString(), r.targetFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating remote target", err.Error())
		return
	}
	plan.Arn = types.StringValue(arn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RemoteTargetRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RemoteTargetRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targets, err := r.client.RustClient.ListRemoteTargets(state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing remote targets", err.Error())
		return
	}
	var found *rustfs.RemoteTarget
	for i := range targets {
		if targets[i].Arn == state.Arn.ValueString() {
			found = &targets[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Arn = types.StringValue(found.Arn)
	state.Endpoint = types.StringValue(found.Endpoint)
	state.Secure = types.BoolValue(found.Secure)
	if found.Region != "" {
		state.Region = types.StringValue(found.Region)
	} else {
		state.Region = types.StringNull()
	}
	if found.Path != "" {
		state.Path = types.StringValue(found.Path)
	} else {
		state.Path = types.StringNull()
	}
	state.TargetBucket = types.StringValue(found.TargetBucket)
	if found.Credentials != nil {
		state.AccessKey = types.StringValue(found.Credentials.AccessKey)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RemoteTargetRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RemoteTargetRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := r.targetFromModel(plan)
	target.Arn = state.Arn.ValueString()
	if _, err := r.client.RustClient.AddRemoteTarget(plan.Bucket.ValueString(), target); err != nil {
		resp.Diagnostics.AddError("Error updating remote target", err.Error())
		return
	}
	plan.Arn = state.Arn
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RemoteTargetRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RemoteTargetRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RustClient.DeleteRemoteTarget(data.Bucket.ValueString(), data.Arn.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting remote target", err.Error())
	}
}

func (r *RemoteTargetRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID is the composite "<source-bucket>:<arn>".
	id := req.ID
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected import ID in the format <source-bucket>:<arn>",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("arn"), parts[1])...)
}

func (r *RemoteTargetRessource) targetFromModel(plan RemoteTargetRessourceModel) rustfs.RemoteTarget {
	return rustfs.RemoteTarget{
		Type:         plan.Type.ValueString(),
		Endpoint:     plan.Endpoint.ValueString(),
		Secure:       plan.Secure.ValueBool(),
		Region:       plan.Region.ValueString(),
		Path:         plan.Path.ValueString(),
		TargetBucket: plan.TargetBucket.ValueString(),
		Credentials: &rustfs.RemoteTargetCredentials{
			AccessKey: plan.AccessKey.ValueString(),
			SecretKey: plan.SecretKey.ValueString(),
		},
	}
}
