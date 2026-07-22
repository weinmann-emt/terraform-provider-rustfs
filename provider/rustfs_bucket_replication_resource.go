package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/minio/minio-go/v7/pkg/replication"
)

var (
	_ resource.Resource                = &BucketReplicationResource{}
	_ resource.ResourceWithImportState = &BucketReplicationResource{}
)

type BucketReplicationResource struct {
	client *AllClient
}

type bucketReplicationResourceModel struct {
	Bucket                 types.String `tfsdk:"bucket"`
	Role                   types.String `tfsdk:"role"`
	DestinationBucket      types.String `tfsdk:"destination_bucket"`
	Priority               types.Int64  `tfsdk:"priority"`
	Status                 types.String `tfsdk:"status"`
	DeleteMarkerReplication types.String `tfsdk:"delete_marker_replication"`
	DeleteReplication      types.String `tfsdk:"delete_replication"`
}

func NewBucketReplicationResource() resource.Resource {
	return &BucketReplicationResource{}
}

func (r *BucketReplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_replication"
}

func (r *BucketReplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS bucket replication",
		MarkdownDescription: "Manage RustFS bucket replication configuration",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the source bucket.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Replication role ARN.",
			},
			"destination_bucket": schema.StringAttribute{
				Required:    true,
				Description: "Destination bucket ARN.",
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Rule priority.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Enabled"),
				Description: "Rule status: Enabled or Disabled.",
			},
			"delete_marker_replication": schema.StringAttribute{
				Optional:    true,
				Description: "Delete marker replication: Enabled or Disabled.",
			},
			"delete_replication": schema.StringAttribute{
				Optional:    true,
				Description: "Delete replication: Enabled or Disabled.",
			},
		},
	}
}

func (r *BucketReplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketReplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketReplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := buildReplicationConfig(plan)
	if err := r.client.Minio.SetBucketReplication(ctx, plan.Bucket.ValueString(), cfg); err != nil {
		resp.Diagnostics.AddError("Error setting bucket replication", "Could not set replication: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketReplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketReplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.Minio.GetBucketReplication(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket replication", "Could not read: "+err.Error())
		return
	}

	state.Role = types.StringValue(cfg.Role)
	if len(cfg.Rules) > 0 {
		rule := cfg.Rules[0]
		state.Status = types.StringValue(string(rule.Status))
		state.Priority = types.Int64Value(int64(rule.Priority))
		state.DeleteMarkerReplication = types.StringValue(string(rule.DeleteMarkerReplication.Status))
		state.DeleteReplication = types.StringValue(string(rule.DeleteReplication.Status))
		if rule.Destination.Bucket != "" {
			state.DestinationBucket = types.StringValue(rule.Destination.Bucket)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketReplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketReplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := buildReplicationConfig(plan)
	if err := r.client.Minio.SetBucketReplication(ctx, plan.Bucket.ValueString(), cfg); err != nil {
		resp.Diagnostics.AddError("Error updating bucket replication", "Could not update: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketReplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketReplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Minio.RemoveBucketReplication(ctx, data.Bucket.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error removing bucket replication", "Could not remove: "+err.Error())
		return
	}
}

func (r *BucketReplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}

func buildReplicationConfig(plan bucketReplicationResourceModel) replication.Config {
	var rules []replication.Rule

	rule := replication.Rule{
		ID:       "rule-1",
		Status:   replication.Status(plan.Status.ValueString()),
		Priority: int(plan.Priority.ValueInt64()),
		Destination: replication.Destination{
			Bucket: plan.DestinationBucket.ValueString(),
		},
	}

	if plan.DeleteMarkerReplication.ValueString() != "" {
		rule.DeleteMarkerReplication = replication.DeleteMarkerReplication{
			Status: replication.Status(plan.DeleteMarkerReplication.ValueString()),
		}
	}
	if plan.DeleteReplication.ValueString() != "" {
		rule.DeleteReplication = replication.DeleteReplication{
			Status: replication.Status(plan.DeleteReplication.ValueString()),
		}
	}

	rules = append(rules, rule)

	return replication.Config{
		Role:  plan.Role.ValueString(),
		Rules: rules,
	}
}
