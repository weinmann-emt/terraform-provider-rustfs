package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &StorageInfoDataSource{}

var storageBackendAttrTypes = map[string]attr.Type{
	"backend_type":         types.StringType,
	"drives_per_set":       types.ListType{ElemType: types.Int64Type},
	"total_sets":           types.ListType{ElemType: types.Int64Type},
	"standard_sc_data":     types.ListType{ElemType: types.Int64Type},
	"standard_sc_parities": types.ListType{ElemType: types.Int64Type},
	"rrsc_data":            types.ListType{ElemType: types.Int64Type},
	"rrsc_parities":        types.ListType{ElemType: types.Int64Type},
}

var storageDiskAttrTypes = map[string]attr.Type{
	"disk_index":    types.Int64Type,
	"endpoint":      types.StringType,
	"avail_space":   types.Int64Type,
	"free_inodes":   types.Int64Type,
	"used_inodes":   types.Int64Type,
	"healing":       types.BoolType,
	"local":         types.BoolType,
	"path":          types.StringType,
	"pool_index":    types.Int64Type,
	"runtime_state": types.StringType,
	"scanning":      types.BoolType,
	"set_index":     types.Int64Type,
	"state":         types.StringType,
	"total_space":   types.Int64Type,
	"used_space":    types.Int64Type,
	"uuid":          types.StringType,
}

type StorageInfoDataSource struct {
	client *AllClient
}

type StorageInfoDataSourceModel struct {
	Backend types.Object `tfsdk:"backend"`
	Disks   types.List   `tfsdk:"disks"`
	RawJSON types.String `tfsdk:"raw_json"`
}

type storageDiskModel struct {
	DiskIndex    types.Int64  `tfsdk:"disk_index"`
	Endpoint     types.String `tfsdk:"endpoint"`
	AvailSpace   types.Int64  `tfsdk:"avail_space"`
	FreeInodes   types.Int64  `tfsdk:"free_inodes"`
	UsedInodes   types.Int64  `tfsdk:"used_inodes"`
	Healing      types.Bool   `tfsdk:"healing"`
	Local        types.Bool   `tfsdk:"local"`
	Path         types.String `tfsdk:"path"`
	PoolIndex    types.Int64  `tfsdk:"pool_index"`
	RuntimeState types.String `tfsdk:"runtime_state"`
	Scanning     types.Bool   `tfsdk:"scanning"`
	SetIndex     types.Int64  `tfsdk:"set_index"`
	State        types.String `tfsdk:"state"`
	TotalSpace   types.Int64  `tfsdk:"total_space"`
	UsedSpace    types.Int64  `tfsdk:"used_space"`
	UUID         types.String `tfsdk:"uuid"`
}

func NewStorageInfoDataSource() datasource.DataSource {
	return &StorageInfoDataSource{}
}

func (d *StorageInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_info"
}

func (d *StorageInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Storage info for the RustFS cluster",
		MarkdownDescription: "Storage layout and health information for the RustFS cluster, including the storage backend and a per-drive breakdown.",
		Attributes: map[string]schema.Attribute{
			"backend": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Storage backend configuration.",
				Attributes: map[string]schema.Attribute{
					"backend_type": schema.StringAttribute{
						Computed:    true,
						Description: "Storage backend type (e.g. Erasure).",
					},
					"drives_per_set": schema.ListAttribute{
						Computed:    true,
						ElementType: types.Int64Type,
						Description: "Number of drives per erasure set.",
					},
					"total_sets": schema.ListAttribute{
						Computed:    true,
						ElementType: types.Int64Type,
						Description: "Total number of erasure sets.",
					},
					"standard_sc_data": schema.ListAttribute{
						Computed:    true,
						ElementType: types.Int64Type,
						Description: "Standard storage class data drives per set.",
					},
					"standard_sc_parities": schema.ListAttribute{
						Computed:    true,
						ElementType: types.Int64Type,
						Description: "Standard storage class parity drives per set.",
					},
					"rrsc_data": schema.ListAttribute{
						Computed:    true,
						ElementType: types.Int64Type,
						Description: "Reduced redundancy storage class data drives per set.",
					},
					"rrsc_parities": schema.ListAttribute{
						Computed:    true,
						ElementType: types.Int64Type,
						Description: "Reduced redundancy storage class parity drives per set.",
					},
				},
			},
			"disks": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Per-drive storage breakdown.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"disk_index":    schema.Int64Attribute{Computed: true, Description: "Index of the disk within the set."},
						"endpoint":      schema.StringAttribute{Computed: true, Description: "Disk endpoint."},
						"avail_space":   schema.Int64Attribute{Computed: true, Description: "Available space in bytes."},
						"free_inodes":   schema.Int64Attribute{Computed: true, Description: "Free inodes."},
						"used_inodes":   schema.Int64Attribute{Computed: true, Description: "Used inodes."},
						"healing":       schema.BoolAttribute{Computed: true, Description: "Whether the drive is healing."},
						"local":         schema.BoolAttribute{Computed: true, Description: "Whether the drive is local."},
						"path":          schema.StringAttribute{Computed: true, Description: "Filesystem path of the drive."},
						"pool_index":    schema.Int64Attribute{Computed: true, Description: "Pool the drive belongs to."},
						"runtime_state": schema.StringAttribute{Computed: true, Description: "Runtime state of the drive."},
						"scanning":      schema.BoolAttribute{Computed: true, Description: "Whether the drive is being scanned."},
						"set_index":     schema.Int64Attribute{Computed: true, Description: "Erasure set the drive belongs to."},
						"state":         schema.StringAttribute{Computed: true, Description: "Health state of the drive."},
						"total_space":   schema.Int64Attribute{Computed: true, Description: "Total space in bytes."},
						"used_space":    schema.Int64Attribute{Computed: true, Description: "Used space in bytes."},
						"uuid":          schema.StringAttribute{Computed: true, Description: "Unique identifier of the drive."},
					},
				},
			},
			"raw_json": schema.StringAttribute{
				Computed:    true,
				Description: "Raw JSON response from the /storageinfo endpoint for full fidelity.",
			},
		},
	}
}

func (d *StorageInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StorageInfoDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	info, err := d.client.RustClient.StorageInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading storage info",
			"Could not read storage info, unexpected error: "+err.Error(),
		)
		return
	}
	if info.Info == nil {
		resp.Diagnostics.AddError(
			"Error reading storage info",
			"Response from storageinfo endpoint was missing the 'info' field.",
		)
		return
	}

	raw, err := json.Marshal(info)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling storage info",
			"Could not marshal storage info: "+err.Error(),
		)
		return
	}

	state := StorageInfoDataSourceModel{
		RawJSON: types.StringValue(string(raw)),
	}

	if info.Info.Backend != nil {
		drivesPerSet, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.DrivesPerSet)
		resp.Diagnostics.Append(diags...)
		totalSets, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.TotalSets)
		resp.Diagnostics.Append(diags...)
		standardSCData, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.StandardSCData)
		resp.Diagnostics.Append(diags...)
		standardSCParities, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.StandardSCParities)
		resp.Diagnostics.Append(diags...)
		rrscData, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.RRSCData)
		resp.Diagnostics.Append(diags...)
		rrscParities, diags := types.ListValueFrom(ctx, types.Int64Type, info.Info.Backend.RRSCParities)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		backendObj, diags := types.ObjectValue(storageBackendAttrTypes, map[string]attr.Value{
			"backend_type":         types.StringValue(info.Info.Backend.BackendType),
			"drives_per_set":       drivesPerSet,
			"total_sets":           totalSets,
			"standard_sc_data":     standardSCData,
			"standard_sc_parities": standardSCParities,
			"rrsc_data":            rrscData,
			"rrsc_parities":        rrscParities,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Backend = backendObj
	}

	disks := make([]storageDiskModel, 0, len(info.Info.Disks))
	for _, disk := range info.Info.Disks {
		disks = append(disks, storageDiskModel{
			DiskIndex:    types.Int64Value(int64(disk.DiskIndex)),
			Endpoint:     types.StringValue(disk.Endpoint),
			AvailSpace:   types.Int64Value(disk.AvailSpace),
			FreeInodes:   types.Int64Value(disk.FreeInodes),
			UsedInodes:   types.Int64Value(disk.UsedInodes),
			Healing:      types.BoolValue(disk.Healing),
			Local:        types.BoolValue(disk.Local),
			Path:         types.StringValue(disk.Path),
			PoolIndex:    types.Int64Value(int64(disk.PoolIndex)),
			RuntimeState: types.StringValue(disk.RuntimeState),
			Scanning:     types.BoolValue(disk.Scanning),
			SetIndex:     types.Int64Value(int64(disk.SetIndex)),
			State:        types.StringValue(disk.State),
			TotalSpace:   types.Int64Value(disk.TotalSpace),
			UsedSpace:    types.Int64Value(disk.UsedSpace),
			UUID:         types.StringValue(disk.UUID),
		})
	}
	diskList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: storageDiskAttrTypes}, disks)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Disks = diskList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
