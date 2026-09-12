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

var _ datasource.DataSource = &HealthInfoDataSource{}

// HealthInfoDataSource provides cluster health and OBD diagnostics.
type HealthInfoDataSource struct {
	client *AllClient
}

// HealthInfoDataSourceModel describes the data source state.
type HealthInfoDataSourceModel struct {
	HealthInfo types.String `tfsdk:"health_info"`
	ObdInfo    types.String `tfsdk:"obd_info"`
	Version    types.String `tfsdk:"version"`
	Region     types.String `tfsdk:"region"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Drives     types.List   `tfsdk:"drives"`
}

type healthDriveModel struct {
	Endpoint       types.String `tfsdk:"endpoint"`
	DrivePath      types.String `tfsdk:"drive_path"`
	State          types.String `tfsdk:"state"`
	TotalSpace     types.Int64  `tfsdk:"total_space"`
	UsedSpace      types.Int64  `tfsdk:"used_space"`
	AvailableSpace types.Int64  `tfsdk:"available_space"`
}

func NewHealthInfoDataSource() datasource.DataSource {
	return &HealthInfoDataSource{}
}

func (d *HealthInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_health_info"
}

func (d *HealthInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Cluster health info and OBD diagnostics for RustFS",
		MarkdownDescription: "Cluster health info and OBD diagnostics for RustFS",
		Attributes: map[string]schema.Attribute{
			"health_info": schema.StringAttribute{
				Computed:            true,
				Description:         "Raw JSON response from the /healthinfo endpoint.",
				MarkdownDescription: "Raw JSON response from the /healthinfo endpoint.",
			},
			"obd_info": schema.StringAttribute{
				Computed:            true,
				Description:         "Raw JSON response from the /obdinfo endpoint.",
				MarkdownDescription: "Raw JSON response from the /obdinfo endpoint.",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				Description:         "RustFS version reported by the health info endpoint.",
				MarkdownDescription: "RustFS version reported by the health info endpoint.",
			},
			"region": schema.StringAttribute{
				Computed:            true,
				Description:         "Region reported by the health info endpoint.",
				MarkdownDescription: "Region reported by the health info endpoint.",
			},
			"timestamp": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp reported by the health info endpoint.",
				MarkdownDescription: "Timestamp reported by the health info endpoint.",
			},
			"drives": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"endpoint": schema.StringAttribute{
							Computed:    true,
							Description: "Drive endpoint.",
						},
						"drive_path": schema.StringAttribute{
							Computed:    true,
							Description: "Drive path.",
						},
						"state": schema.StringAttribute{
							Computed:    true,
							Description: "Drive state (for example ok, degraded, offline).",
						},
						"total_space": schema.Int64Attribute{
							Computed:    true,
							Description: "Total drive space in bytes.",
						},
						"used_space": schema.Int64Attribute{
							Computed:    true,
							Description: "Used drive space in bytes.",
						},
						"available_space": schema.Int64Attribute{
							Computed:    true,
							Description: "Available drive space in bytes.",
						},
					},
				},
				Description: "List of drives and their health states.",
			},
		},
	}
}

func (d *HealthInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AllClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AllClient, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *HealthInfoDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	health, err := d.client.RustClient.GetHealthInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading health info",
			"Could not read health info: "+err.Error(),
		)
		return
	}

	obd, err := d.client.RustClient.GetObdInfo()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading OBD info",
			"Could not read OBD info: "+err.Error(),
		)
		return
	}

	healthRaw, err := json.Marshal(health)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling health info",
			"Could not marshal health info: "+err.Error(),
		)
		return
	}
	obdRaw, err := json.Marshal(obd)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling OBD info",
			"Could not marshal OBD info: "+err.Error(),
		)
		return
	}

	drives := make([]healthDriveModel, 0, len(health.Drives))
	for _, drv := range health.Drives {
		drives = append(drives, healthDriveModel{
			Endpoint:       types.StringValue(drv.Endpoint),
			DrivePath:      types.StringValue(drv.DrivePath),
			State:          types.StringValue(drv.State),
			TotalSpace:     types.Int64Value(drv.TotalSpace),
			UsedSpace:      types.Int64Value(drv.UsedSpace),
			AvailableSpace: types.Int64Value(drv.AvailableSpace),
		})
	}
	driveList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"endpoint":        types.StringType,
			"drive_path":      types.StringType,
			"state":           types.StringType,
			"total_space":     types.Int64Type,
			"used_space":      types.Int64Type,
			"available_space": types.Int64Type,
		},
	}, drives)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := HealthInfoDataSourceModel{
		HealthInfo: types.StringValue(string(healthRaw)),
		ObdInfo:    types.StringValue(string(obdRaw)),
		Version:    types.StringValue(health.Version),
		Region:     types.StringValue(health.Region),
		Timestamp:  types.StringValue(health.Timestamp),
		Drives:     driveList,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
