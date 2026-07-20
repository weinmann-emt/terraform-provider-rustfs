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
	"github.com/minio/minio-go/v7/pkg/notification"
)

var (
	_ resource.Resource                = &BucketNotificationResource{}
	_ resource.ResourceWithImportState = &BucketNotificationResource{}
)

type BucketNotificationResource struct {
	client *AllClient
}

type bucketNotificationQueueModel struct {
	Arn          types.String `tfsdk:"arn"`
	Events       types.Set    `tfsdk:"events"`
	FilterPrefix types.String `tfsdk:"filter_prefix"`
	FilterSuffix types.String `tfsdk:"filter_suffix"`
}

type bucketNotificationResourceModel struct {
	Bucket types.String                   `tfsdk:"bucket"`
	Queue  []bucketNotificationQueueModel `tfsdk:"queue"`
}

func NewBucketNotificationResource() resource.Resource {
	return &BucketNotificationResource{}
}

func (r *BucketNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_notification"
}

func (r *BucketNotificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage RustFS bucket event notifications",
		MarkdownDescription: "Manage RustFS bucket event notification configuration",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"queue": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Queue notification configurations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"arn": schema.StringAttribute{
							Required:    true,
							Description: "ARN of the queue target (e.g., arn:minio:sqs::PRIMARY:amqp).",
						},
						"events": schema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "S3 event types (e.g., s3:ObjectCreated:*, s3:ObjectRemoved:*).",
						},
						"filter_prefix": schema.StringAttribute{
							Optional:    true,
							Description: "Filter events by object key prefix.",
						},
						"filter_suffix": schema.StringAttribute{
							Optional:    true,
							Description: "Filter events by object key suffix.",
						},
					},
				},
			},
		},
	}
}

func (r *BucketNotificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := buildNotificationConfig(plan)
	if err := r.client.Minio.SetBucketNotification(ctx, plan.Bucket.ValueString(), config); err != nil {
		resp.Diagnostics.AddError(
			"Error setting bucket notification",
			"Could not set bucket notification: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.Minio.GetBucketNotification(ctx, state.Bucket.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket notification",
			"Could not read bucket notification: "+err.Error(),
		)
		return
	}

	var queues []bucketNotificationQueueModel
	for _, q := range config.QueueConfigs {
		var events []string
		for _, e := range q.Events {
			events = append(events, string(e))
		}
		eventsSet, diags := types.SetValueFrom(ctx, types.StringType, events)
		resp.Diagnostics.Append(diags...)

		var prefix, suffix string
		if q.Filter != nil {
			for _, rule := range q.Filter.S3Key.FilterRules {
				switch rule.Name {
				case "prefix":
					prefix = rule.Value
				case "suffix":
					suffix = rule.Value
				}
			}
		}

		queues = append(queues, bucketNotificationQueueModel{
			Arn:          types.StringValue(q.Queue),
			Events:       eventsSet,
			FilterPrefix: types.StringValue(prefix),
			FilterSuffix: types.StringValue(suffix),
		})
	}
	state.Queue = queues

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := buildNotificationConfig(plan)
	if err := r.client.Minio.SetBucketNotification(ctx, plan.Bucket.ValueString(), config); err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket notification",
			"Could not update bucket notification: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Minio.RemoveAllBucketNotification(ctx, data.Bucket.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error removing bucket notification",
			"Could not remove bucket notification: "+err.Error(),
		)
		return
	}
}

func (r *BucketNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}

func buildNotificationConfig(plan bucketNotificationResourceModel) notification.Configuration {
	var config notification.Configuration
	for _, q := range plan.Queue {
		var events []notification.EventType
		q.Events.ElementsAs(nil, &events, false)

		filter := &notification.Filter{}
		if prefix := q.FilterPrefix.ValueString(); prefix != "" {
			filter.S3Key.FilterRules = append(filter.S3Key.FilterRules,
				notification.FilterRule{Name: "prefix", Value: prefix})
		}
		if suffix := q.FilterSuffix.ValueString(); suffix != "" {
			filter.S3Key.FilterRules = append(filter.S3Key.FilterRules,
				notification.FilterRule{Name: "suffix", Value: suffix})
		}
		if len(filter.S3Key.FilterRules) == 0 {
			filter = nil
		}

		config.QueueConfigs = append(config.QueueConfigs, notification.QueueConfig{
			Config: notification.Config{
				Events: events,
				Filter: filter,
			},
			Queue: q.Arn.ValueString(),
		})
	}
	return config
}
