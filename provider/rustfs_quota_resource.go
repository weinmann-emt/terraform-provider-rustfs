package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

const (
	// quotaReadMaxAttempts and quotaReadRetryInterval bound the time the quota
	// read tolerates RustFS returning ServiceUnavailable while the scanner is
	// still computing the bucket's authoritative usage on a freshly started
	// server (usually available within tens of seconds).
	quotaReadMaxAttempts   = 30
	quotaReadRetryInterval = 3 * time.Second
)

func isTransientQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if !strings.Contains(msg, "ServiceUnavailable") {
		return false
	}
	return strings.Contains(msg, "authoritative bucket usage") ||
		strings.Contains(msg, "durable quota capability is not confirmed")
}

// quotaReadWithRetry retries the quota read while the server reports that the
// bucket's authoritative usage is not computed yet. Other errors fail fast.
func quotaReadWithRetry(ctx context.Context, bucket string, read func(string) (rustfs.Quota, error)) (rustfs.Quota, error) {
	var lastErr error
	for attempt := 0; attempt < quotaReadMaxAttempts; attempt++ {
		quota, err := read(bucket)
		if err == nil {
			return quota, nil
		}
		lastErr = err
		if !isTransientQuotaError(err) {
			return rustfs.Quota{}, err
		}
		timer := time.NewTimer(quotaReadRetryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return rustfs.Quota{}, ctx.Err()
		case <-timer.C:
		}
	}
	return rustfs.Quota{}, lastErr
}

// quotaSetWithRetry retries setting the bucket quota while the cluster has not
// yet confirmed the durable quota capability (fresh single-node startup).
// Other errors fail fast.
func quotaSetWithRetry(ctx context.Context, quota rustfs.Quota, set func(rustfs.Quota) (rustfs.Quota, error)) (rustfs.Quota, error) {
	var lastErr error
	for attempt := 0; attempt < quotaReadMaxAttempts; attempt++ {
		got, err := set(quota)
		if err == nil {
			return got, nil
		}
		lastErr = err
		if !isTransientQuotaError(err) {
			return rustfs.Quota{}, err
		}
		timer := time.NewTimer(quotaReadRetryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return rustfs.Quota{}, ctx.Err()
		case <-timer.C:
		}
	}
	return rustfs.Quota{}, lastErr
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &quotaResource{}
	_ resource.ResourceWithImportState = &quotaResource{}
)

// NewquotaResource is a helper function to simplify the provider implementation.
func NewquotaResource() resource.Resource {
	return &quotaResource{}
}

// quotaResource is the resource implementation.
type quotaResource struct {
	client *AllClient
}

type quotaResourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Quota  types.Int64  `tfsdk:"quota"`
}

// Metadata returns the resource type name.
func (r *quotaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota"
}

// Schema defines the schema for the resource.
func (r *quotaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage buckets quota in rustfs",
		MarkdownDescription: "Manage bucket quota in rustfs",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket",
			},
			"quota": schema.Int64Attribute{
				Required:    true,
				Description: "Bytes of the quota",
			},
		},
	}
}

func (r *quotaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *quotaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan quotaResourceModel
	diags := req.Plan.Get(ctx, &plan)
	// ToDo: Check if bucket exists
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	q := rustfs.Quota{Bucket: plan.Bucket.ValueString(), Quota: int(plan.Quota.ValueInt64()), Quota_Type: "HARD"}
	_, err := quotaSetWithRetry(ctx, q, r.client.RustClient.SetQuota)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket quota",
			"Could not create bucket quota, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *quotaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state quotaResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Read
	read, err := quotaReadWithRetry(ctx, state.Bucket.ValueString(), r.client.RustClient.ReadQuota)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket quota",
			"Could not read bucket quota, unexpected error: "+err.Error(),
		)
		return
	}
	// Save update status
	state.Quota = types.Int64Value(int64(read.Quota))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *quotaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan quotaResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	quota := rustfs.Quota{Bucket: plan.Bucket.ValueString(), Quota: int(plan.Quota.ValueInt64()), Quota_Type: "HARD"}
	read, err := quotaSetWithRetry(ctx, quota, r.client.RustClient.SetQuota)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket quota",
			"Could not update bucket quota, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Quota = types.Int64Value(int64(read.Quota))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *quotaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data quotaResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.DeletQuota(data.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bucket quota",
			"Could not delete bucket quota, unexpected error: "+err.Error(),
		)
	}
}

func (r *quotaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}
