package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var _ datasource.DataSource = &ServerInfoDataSource{}

type ServerInfoDataSource struct {
	client *AllClient
}

type ServerInfoDataSourceModel struct {
	Mode              types.String `tfsdk:"mode"`
	DeploymentID      types.String `tfsdk:"deployment_id"`
	Region            types.String `tfsdk:"region"`
	BitrotSelftest    types.String `tfsdk:"bitrot_selftest"`
	BackendType       types.String `tfsdk:"backend_type"`
	OfflineDisks      types.Int64  `tfsdk:"offline_disks"`
	OnlineDisks       types.Int64  `tfsdk:"online_disks"`
	TotalDrivesPerSet types.List   `tfsdk:"total_drives_per_set"`
	TotalSets         types.List   `tfsdk:"total_sets"`
	BucketCount       types.Int64  `tfsdk:"bucket_count"`
	ObjectCount       types.Int64  `tfsdk:"object_count"`
	VersionCount      types.Int64  `tfsdk:"version_count"`
	DeleteMarkerCount types.Int64  `tfsdk:"delete_marker_count"`
	UsageSize         types.Int64  `tfsdk:"usage_size"`
	PoolCount         types.Int64  `tfsdk:"pool_count"`
	Pools             types.List   `tfsdk:"pools"`
	Servers           types.List   `tfsdk:"servers"`
	RawJSON           types.String `tfsdk:"raw_json"`
}

type serverInfoPoolModel struct {
	PoolNumber         types.Int64 `tfsdk:"pool_number"`
	SetNumber          types.Int64 `tfsdk:"set_number"`
	ID                 types.Int64 `tfsdk:"id"`
	RawCapacity        types.Int64 `tfsdk:"raw_capacity"`
	RawUsage           types.Int64 `tfsdk:"raw_usage"`
	Usage              types.Int64 `tfsdk:"usage"`
	ObjectsCount       types.Int64 `tfsdk:"objects_count"`
	VersionsCount      types.Int64 `tfsdk:"versions_count"`
	DeleteMarkersCount types.Int64 `tfsdk:"delete_markers_count"`
	HealDisks          types.Int64 `tfsdk:"heal_disks"`
}

type serverInfoDriveModel struct {
	Endpoint     types.String  `tfsdk:"endpoint"`
	Path         types.String  `tfsdk:"path"`
	State        types.String  `tfsdk:"state"`
	RuntimeState types.String  `tfsdk:"runtime_state"`
	Healing      types.Bool    `tfsdk:"healing"`
	Local        types.Bool    `tfsdk:"local"`
	UUID         types.String  `tfsdk:"uuid"`
	Totalspace   types.Int64   `tfsdk:"totalspace"`
	Usedspace    types.Int64   `tfsdk:"usedspace"`
	Availspace   types.Int64   `tfsdk:"availspace"`
	Utilization  types.Float64 `tfsdk:"utilization"`
}

type serverInfoServerModel struct {
	Endpoint      types.String           `tfsdk:"endpoint"`
	State         types.String           `tfsdk:"state"`
	Version       types.String           `tfsdk:"version"`
	Uptime        types.Int64            `tfsdk:"uptime"`
	NumCPU        types.Int64            `tfsdk:"num_cpu"`
	MaxProcs      types.Int64            `tfsdk:"max_procs"`
	MemAlloc      types.Int64            `tfsdk:"mem_alloc"`
	MemTotalAlloc types.Int64            `tfsdk:"mem_total_alloc"`
	Drives        []serverInfoDriveModel `tfsdk:"drives"`
}

func NewServerInfoDataSource() datasource.DataSource {
	return &ServerInfoDataSource{}
}

func (d *ServerInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_info"
}

func (d *ServerInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Server info for RustFS cluster",
		MarkdownDescription: "Exposes cluster, server, pool and drive information from the RustFS admin API.",
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				Computed:    true,
				Description: "Cluster mode (e.g. online, offline).",
			},
			"deployment_id": schema.StringAttribute{
				Computed:    true,
				Description: "Cluster deployment ID.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "Cluster region, when configured.",
			},
			"bitrot_selftest": schema.StringAttribute{
				Computed:    true,
				Description: "Bitrot self test status.",
			},
			"backend_type": schema.StringAttribute{
				Computed:    true,
				Description: "Backend type (e.g. Erasure).",
			},
			"offline_disks": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of offline disks.",
			},
			"online_disks": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of online disks.",
			},
			"total_drives_per_set": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Total drives per erasure set.",
			},
			"total_sets": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Total number of erasure sets per pool.",
			},
			"bucket_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of buckets in the cluster.",
			},
			"object_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of objects in the cluster.",
			},
			"version_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of object versions in the cluster.",
			},
			"delete_marker_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of delete markers in the cluster.",
			},
			"usage_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Aggregate usage size in bytes.",
			},
			"pool_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of storage pools in the cluster.",
			},
			"pools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pool_number": schema.Int64Attribute{
							Computed: true,
						},
						"set_number": schema.Int64Attribute{
							Computed: true,
						},
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"raw_capacity": schema.Int64Attribute{
							Computed: true,
						},
						"raw_usage": schema.Int64Attribute{
							Computed: true,
						},
						"usage": schema.Int64Attribute{
							Computed: true,
						},
						"objects_count": schema.Int64Attribute{
							Computed: true,
						},
						"versions_count": schema.Int64Attribute{
							Computed: true,
						},
						"delete_markers_count": schema.Int64Attribute{
							Computed: true,
						},
						"heal_disks": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
				Description: "Storage pools and their erasure sets.",
			},
			"servers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"endpoint": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"version": schema.StringAttribute{
							Computed: true,
						},
						"uptime": schema.Int64Attribute{
							Computed: true,
						},
						"num_cpu": schema.Int64Attribute{
							Computed: true,
						},
						"max_procs": schema.Int64Attribute{
							Computed: true,
						},
						"mem_alloc": schema.Int64Attribute{
							Computed: true,
						},
						"mem_total_alloc": schema.Int64Attribute{
							Computed: true,
						},
						"drives": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"endpoint": schema.StringAttribute{
										Computed: true,
									},
									"path": schema.StringAttribute{
										Computed: true,
									},
									"state": schema.StringAttribute{
										Computed: true,
									},
									"runtime_state": schema.StringAttribute{
										Computed: true,
									},
									"healing": schema.BoolAttribute{
										Computed: true,
									},
									"local": schema.BoolAttribute{
										Computed: true,
									},
									"uuid": schema.StringAttribute{
										Computed: true,
									},
									"totalspace": schema.Int64Attribute{
										Computed: true,
									},
									"usedspace": schema.Int64Attribute{
										Computed: true,
									},
									"availspace": schema.Int64Attribute{
										Computed: true,
									},
									"utilization": schema.Float64Attribute{
										Computed: true,
									},
								},
							},
							Description: "Drives attached to this server.",
						},
					},
				},
				Description: "Server nodes in the cluster.",
			},
			"raw_json": schema.StringAttribute{
				Computed:            true,
				Description:         "Full raw JSON response from the /info endpoint.",
				MarkdownDescription: "Full raw JSON response from the /info endpoint.",
			},
		},
	}
}

func (d *ServerInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerInfoDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	info, err := d.client.RustClient.ServerInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading server info",
			"Could not read server info, unexpected error: "+err.Error(),
		)
		return
	}

	state := ServerInfoDataSourceModel{
		Mode:              types.StringValue(info.Info.Mode),
		DeploymentID:      types.StringValue(info.Info.DeploymentID),
		Region:            optionalString(info.Info.Region),
		BitrotSelftest:    types.StringValue(info.BitrotSelftest),
		BackendType:       types.StringValue(info.Info.Backend.BackendType),
		OfflineDisks:      types.Int64Value(info.Info.Backend.OfflineDisks),
		OnlineDisks:       types.Int64Value(info.Info.Backend.OnlineDisks),
		BucketCount:       types.Int64Value(info.Info.Buckets.Count),
		ObjectCount:       types.Int64Value(info.Info.Objects.Count),
		VersionCount:      types.Int64Value(info.Info.Versions.Count),
		DeleteMarkerCount: types.Int64Value(info.Info.Deletemarkers.Count),
		UsageSize:         types.Int64Value(info.Info.Usage.Size),
	}

	drivesPerSet, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.TotalDrivesPerSet)
	resp.Diagnostics.Append(diags...)
	totalSets, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.TotalSets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.TotalDrivesPerSet = drivesPerSet
	state.TotalSets = totalSets

	poolModels := flattenPools(info.Info.Pools)
	state.PoolCount = types.Int64Value(int64(len(poolModels)))
	pools, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: poolObjectType(),
	}, poolModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Pools = pools

	serverModels := flattenServers(info.Info.Servers)
	servers, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: serverObjectType(),
	}, serverModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Servers = servers

	raw, err := json.Marshal(info)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling server info",
			"Could not marshal server info: "+err.Error(),
		)
		return
	}
	state.RawJSON = types.StringValue(string(raw))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func optionalString(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func poolObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"pool_number":          types.Int64Type,
		"set_number":           types.Int64Type,
		"id":                   types.Int64Type,
		"raw_capacity":         types.Int64Type,
		"raw_usage":            types.Int64Type,
		"usage":                types.Int64Type,
		"objects_count":        types.Int64Type,
		"versions_count":       types.Int64Type,
		"delete_markers_count": types.Int64Type,
		"heal_disks":           types.Int64Type,
	}
}

func serverObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"endpoint":        types.StringType,
		"state":           types.StringType,
		"version":         types.StringType,
		"uptime":          types.Int64Type,
		"num_cpu":         types.Int64Type,
		"max_procs":       types.Int64Type,
		"mem_alloc":       types.Int64Type,
		"mem_total_alloc": types.Int64Type,
		"drives":          types.ListType{ElemType: types.ObjectType{AttrTypes: driveObjectType()}},
	}
}

func driveObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"endpoint":      types.StringType,
		"path":          types.StringType,
		"state":         types.StringType,
		"runtime_state": types.StringType,
		"healing":       types.BoolType,
		"local":         types.BoolType,
		"uuid":          types.StringType,
		"totalspace":    types.Int64Type,
		"usedspace":     types.Int64Type,
		"availspace":    types.Int64Type,
		"utilization":   types.Float64Type,
	}
}

func flattenPools(pools map[string]map[string]rustfs.PoolSetInfo) []serverInfoPoolModel {
	var out []serverInfoPoolModel
	var poolNumbers []int
	for poolNumber := range pools {
		n, err := strconv.Atoi(poolNumber)
		if err != nil {
			continue
		}
		poolNumbers = append(poolNumbers, n)
	}
	sort.Ints(poolNumbers)
	for _, poolNumber := range poolNumbers {
		sets := pools[strconv.Itoa(poolNumber)]
		var setNumbers []int
		for setNumber := range sets {
			n, err := strconv.Atoi(setNumber)
			if err != nil {
				continue
			}
			setNumbers = append(setNumbers, n)
		}
		sort.Ints(setNumbers)
		for _, setNumber := range setNumbers {
			p := sets[strconv.Itoa(setNumber)]
			out = append(out, serverInfoPoolModel{
				PoolNumber:         types.Int64Value(int64(poolNumber)),
				SetNumber:          types.Int64Value(int64(setNumber)),
				ID:                 types.Int64Value(p.ID),
				RawCapacity:        types.Int64Value(p.RawCapacity),
				RawUsage:           types.Int64Value(p.RawUsage),
				Usage:              types.Int64Value(p.Usage),
				ObjectsCount:       types.Int64Value(p.ObjectsCount),
				VersionsCount:      types.Int64Value(p.VersionsCount),
				DeleteMarkersCount: types.Int64Value(p.DeleteMarkersCount),
				HealDisks:          types.Int64Value(p.HealDisks),
			})
		}
	}
	return out
}

func flattenServers(servers []rustfs.ServerEntry) []serverInfoServerModel {
	out := make([]serverInfoServerModel, 0, len(servers))
	for _, s := range servers {
		drives := make([]serverInfoDriveModel, 0, len(s.Drives))
		for _, dr := range s.Drives {
			drives = append(drives, serverInfoDriveModel{
				Endpoint:     types.StringValue(dr.Endpoint),
				Path:         types.StringValue(dr.Path),
				State:        types.StringValue(dr.State),
				RuntimeState: types.StringValue(dr.RuntimeState),
				Healing:      types.BoolValue(dr.Healing),
				Local:        types.BoolValue(dr.Local),
				UUID:         types.StringValue(dr.UUID),
				Totalspace:   types.Int64Value(dr.Totalspace),
				Usedspace:    types.Int64Value(dr.Usedspace),
				Availspace:   types.Int64Value(dr.Availspace),
				Utilization:  types.Float64Value(dr.Utilization),
			})
		}
		out = append(out, serverInfoServerModel{
			Endpoint: types.StringValue(s.Endpoint),
			State:    types.StringValue(s.State),
			Version:  types.StringValue(s.Version),
			Uptime:   types.Int64Value(s.Uptime),
			NumCPU:   types.Int64Value(int64(s.NumCPU)),
			MaxProcs: types.Int64Value(int64(s.MaxProcs)),
			//#nosec G115 — memory stats are byte counts that fit in int64
			MemAlloc: types.Int64Value(int64(s.MemStats.Alloc)),
			//#nosec G115 — memory stats are byte counts that fit in int64
			MemTotalAlloc: types.Int64Value(int64(s.MemStats.TotalAlloc)),
			Drives:        drives,
		})
	}
	return out
}
