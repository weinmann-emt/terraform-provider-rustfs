package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var _ datasource.DataSource = &ReplicationMetricsDataSource{}

type ReplicationMetricsDataSource struct {
	client *AllClient
}

type ReplicationMetricsDataSourceModel struct {
	Bucket                   types.String `tfsdk:"bucket"`
	ID                       types.String `tfsdk:"id"`
	ReplicationCount         types.Int64  `tfsdk:"replication_count"`
	CompletedReplicationSize types.Int64  `tfsdk:"completed_replication_size"`
	ReplicaCount             types.Int64  `tfsdk:"replica_count"`
	ReplicaSize              types.Int64  `tfsdk:"replica_size"`
	Failed                   types.Object `tfsdk:"failed"`
	Queued                   types.Object `tfsdk:"queued"`
	Targets                  types.List   `tfsdk:"targets"`
	JSON                     types.String `tfsdk:"json"`
}

var replicationRStatAttrTypes = map[string]attr.Type{
	"count": types.Float64Type,
	"bytes": types.Int64Type,
}

var replicationTimedErrStatsAttrTypes = map[string]attr.Type{
	"last_minute": types.ObjectType{AttrTypes: replicationRStatAttrTypes},
	"last_hour":   types.ObjectType{AttrTypes: replicationRStatAttrTypes},
	"totals":      types.ObjectType{AttrTypes: replicationRStatAttrTypes},
}

var replicationQStatAttrTypes = map[string]attr.Type{
	"count": types.Float64Type,
	"bytes": types.Float64Type,
}

var replicationInQueueMetricAttrTypes = map[string]attr.Type{
	"curr": types.ObjectType{AttrTypes: replicationQStatAttrTypes},
	"avg":  types.ObjectType{AttrTypes: replicationQStatAttrTypes},
	"max":  types.ObjectType{AttrTypes: replicationQStatAttrTypes},
	"peak": types.ObjectType{AttrTypes: replicationQStatAttrTypes},
}

var replicationTargetMetricAttrTypes = map[string]attr.Type{
	"target":                     types.StringType,
	"replication_count":          types.Int64Type,
	"completed_replication_size": types.Int64Type,
	"limit_in_bits":              types.Int64Type,
	"current_bandwidth":          types.Float64Type,
	"failed":                     types.ObjectType{AttrTypes: replicationTimedErrStatsAttrTypes},
	"failed_replication_size":    types.Int64Type,
	"failed_replication_count":   types.Int64Type,
}

func NewReplicationMetricsDataSource() datasource.DataSource {
	return &ReplicationMetricsDataSource{}
}

func (d *ReplicationMetricsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_replication_metrics"
}

func (d *ReplicationMetricsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Expose replication metrics for a bucket",
		MarkdownDescription: "Exposes replication transfer metrics (counts, bytes, queue backlog) for a bucket from RustFS.",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Bucket to read replication metrics for.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"replication_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Total number of objects replicated for the bucket.",
			},
			"completed_replication_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Total number of bytes replicated for the bucket.",
			},
			"replica_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Total number of replicas created for the bucket.",
			},
			"replica_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Total size in bytes of all replicas created for the bucket.",
			},
			"failed": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Aggregate replication failure counters.",
				Attributes:  replicationTimedErrStatsSchema(),
			},
			"queued": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Replication queue backlog.",
				Attributes:  replicationInQueueMetricSchema(),
			},
			"targets": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Per-target replication transfer statistics.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: replicationTargetMetricSchema(),
				},
			},
			"json": schema.StringAttribute{
				Computed:    true,
				Description: "Raw JSON response as returned by the RustFS admin API.",
			},
		},
	}
}

func replicationRStatSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"count": schema.Float64Attribute{
			Computed:    true,
			Description: "Number of objects.",
		},
		"bytes": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of bytes.",
		},
	}
}

func replicationTimedErrStatsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"last_minute": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Failures in the last minute.",
			Attributes:  replicationRStatSchema(),
		},
		"last_hour": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Failures in the last hour.",
			Attributes:  replicationRStatSchema(),
		},
		"totals": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Cumulative failures since uptime.",
			Attributes:  replicationRStatSchema(),
		},
	}
}

func replicationInQueueMetricSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"curr": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Current queue backlog.",
			Attributes:  replicationQStatSchema(),
		},
		"avg": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Average queue backlog.",
			Attributes:  replicationQStatSchema(),
		},
		"max": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Maximum queue backlog since uptime.",
			Attributes:  replicationQStatSchema(),
		},
		"peak": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Peak queue backlog since uptime.",
			Attributes:  replicationQStatSchema(),
		},
	}
}

func replicationQStatSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"count": schema.Float64Attribute{
			Computed:    true,
			Description: "Number of queued objects.",
		},
		"bytes": schema.Float64Attribute{
			Computed:    true,
			Description: "Number of queued bytes.",
		},
	}
}

func replicationTargetMetricSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"target": schema.StringAttribute{
			Computed:    true,
			Description: "Replication target ARN.",
		},
		"replication_count": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of objects replicated to this target.",
		},
		"completed_replication_size": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of bytes replicated to this target.",
		},
		"limit_in_bits": schema.Int64Attribute{
			Computed:    true,
			Description: "Bandwidth limit for this target in bytes/sec.",
		},
		"current_bandwidth": schema.Float64Attribute{
			Computed:    true,
			Description: "Current replication bandwidth for this target in bytes/sec.",
		},
		"failed": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Replication failure counters for this target.",
			Attributes:  replicationTimedErrStatsSchema(),
		},
		"failed_replication_size": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of bytes that failed to replicate to this target.",
		},
		"failed_replication_count": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of objects that failed to replicate to this target.",
		},
	}
}

func (d *ReplicationMetricsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AllClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *ReplicationMetricsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ReplicationMetricsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket := config.Bucket.ValueString()
	if bucket == "" {
		resp.Diagnostics.AddError(
			"Error reading replication metrics",
			"bucket is required",
		)
		return
	}

	metrics, err := d.client.RustClient.ReplicationMetrics(bucket)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading replication metrics",
			"Could not read replication metrics: "+err.Error(),
		)
		return
	}

	config.Bucket = types.StringValue(bucket)
	config.ID = types.StringValue("replicationmetrics/" + bucket)
	config.ReplicationCount = types.Int64Value(metrics.ReplicationCount)
	config.CompletedReplicationSize = types.Int64Value(metrics.CompletedReplicationSize)
	config.ReplicaCount = types.Int64Value(metrics.ReplicaCount)
	config.ReplicaSize = types.Int64Value(metrics.ReplicaSize)
	config.Failed = replicationTimedErrStatsValue(metrics.Failed)
	config.Queued = replicationInQueueMetricValue(metrics.Queued)
	config.Targets = replicationTargetsValue(metrics.Stats)

	if len(metrics.Raw) > 0 {
		config.JSON = types.StringValue(string(metrics.Raw))
	} else {
		config.JSON = types.StringValue("{}")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func replicationRStatValue(stat rustfs.ReplicationRStat) types.Object {
	return types.ObjectValueMust(replicationRStatAttrTypes, map[string]attr.Value{
		"count": types.Float64Value(stat.Count),
		"bytes": types.Int64Value(stat.Bytes),
	})
}

func replicationTimedErrStatsValue(stats rustfs.ReplicationTimedErrStats) types.Object {
	return types.ObjectValueMust(replicationTimedErrStatsAttrTypes, map[string]attr.Value{
		"last_minute": replicationRStatValue(stats.LastMinute),
		"last_hour":   replicationRStatValue(stats.LastHour),
		"totals":      replicationRStatValue(stats.Totals),
	})
}

func replicationQStatValue(stat rustfs.ReplicationQStat) types.Object {
	return types.ObjectValueMust(replicationQStatAttrTypes, map[string]attr.Value{
		"count": types.Float64Value(stat.Count),
		"bytes": types.Float64Value(stat.Bytes),
	})
}

func replicationInQueueMetricValue(metric rustfs.ReplicationInQueueMetric) types.Object {
	return types.ObjectValueMust(replicationInQueueMetricAttrTypes, map[string]attr.Value{
		"curr": replicationQStatValue(metric.Curr),
		"avg":  replicationQStatValue(metric.Avg),
		"max":  replicationQStatValue(metric.Max),
		"peak": replicationQStatValue(metric.Peak),
	})
}

func replicationTargetMetricValue(target string, metric rustfs.ReplicationTargetMetric) types.Object {
	return types.ObjectValueMust(replicationTargetMetricAttrTypes, map[string]attr.Value{
		"target":                     types.StringValue(target),
		"replication_count":          types.Int64Value(metric.ReplicationCount),
		"completed_replication_size": types.Int64Value(metric.CompletedReplicationSize),
		"limit_in_bits":              types.Int64Value(metric.LimitInBits),
		"current_bandwidth":          types.Float64Value(metric.CurrentBandwidth),
		"failed":                     replicationTimedErrStatsValue(metric.Failed),
		"failed_replication_size":    types.Int64Value(metric.FailedReplicationSize),
		"failed_replication_count":   types.Int64Value(metric.FailedReplicationCount),
	})
}

func replicationTargetsValue(stats map[string]rustfs.ReplicationTargetMetric) types.List {
	targets := make([]string, 0, len(stats))
	for arn := range stats {
		targets = append(targets, arn)
	}
	sort.Strings(targets)

	values := make([]attr.Value, 0, len(targets))
	for _, arn := range targets {
		values = append(values, replicationTargetMetricValue(arn, stats[arn]))
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: replicationTargetMetricAttrTypes}, values)
}
