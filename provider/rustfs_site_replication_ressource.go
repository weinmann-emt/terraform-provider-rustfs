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
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

var (
	_ resource.Resource                = &SiteReplicationRessource{}
	_ resource.ResourceWithImportState = &SiteReplicationRessource{}
)

func NewSiteReplicationRessource() resource.Resource {
	return &SiteReplicationRessource{}
}

type SiteReplicationRessource struct {
	client *AllClient
}

type SiteReplicationRessourceModel struct {
	Name          types.String `tfsdk:"name"`
	Endpoint      types.String `tfsdk:"endpoint"`
	AccessKey     types.String `tfsdk:"access_key"`
	SecretKey     types.String `tfsdk:"secret_key"`
	SkipTLSVerify types.Bool   `tfsdk:"skip_tls_verify"`
	CACertPEM     types.String `tfsdk:"ca_cert_pem"`
}

func (r *SiteReplicationRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_replication"
}

func (r *SiteReplicationRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage site-replication peers in rustfs",
		MarkdownDescription: "Manage site-replication peers in rustfs",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name of the replication peer. Changing this forces recreation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Endpoint of the peer site",
			},
			"access_key": schema.StringAttribute{
				Required:    true,
				Description: "Access key of the peer site",
			},
			"secret_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Secret key of the peer site",
			},
			"skip_tls_verify": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Skip TLS certificate verification when connecting to the peer. Defaults to false.",
			},
			"ca_cert_pem": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Custom CA certificate (PEM) for the peer connection",
			},
		},
	}
}

func (r *SiteReplicationRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SiteReplicationRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SiteReplicationRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := modelToSite(plan)
	if err := r.client.RustClient.SiteReplicationAdd(site); err != nil {
		resp.Diagnostics.AddError("Error creating site replication", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteReplicationRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SiteReplicationRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	info, err := r.client.RustClient.SiteReplicationInfo()
	if err != nil {
		resp.Diagnostics.AddError("Error reading site replication", err.Error())
		return
	}
	var found *rustfs.SiteReplicationPeer
	for i := range info.Sites {
		if info.Sites[i].Name == state.Name.ValueString() {
			found = &info.Sites[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Endpoint = types.StringValue(found.Endpoint)
	state.SkipTLSVerify = types.BoolValue(found.SkipTLSVerify)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SiteReplicationRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SiteReplicationRessourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := modelToSite(plan)
	if err := r.client.RustClient.SiteReplicationEdit(site); err != nil {
		resp.Diagnostics.AddError("Error updating site replication", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteReplicationRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SiteReplicationRessourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RustClient.SiteReplicationRemove([]string{data.Name.ValueString()}); err != nil {
		resp.Diagnostics.AddError("Error deleting site replication", err.Error())
	}
}

func (r *SiteReplicationRessource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func modelToSite(plan SiteReplicationRessourceModel) rustfs.SiteReplicationSite {
	return rustfs.SiteReplicationSite{
		Name:          plan.Name.ValueString(),
		Endpoint:      strings.TrimSuffix(plan.Endpoint.ValueString(), "/"),
		AccessKey:     plan.AccessKey.ValueString(),
		SecretKey:     plan.SecretKey.ValueString(),
		SkipTLSVerify: plan.SkipTLSVerify.ValueBool(),
		CACertPEM:     plan.CACertPEM.ValueString(),
	}
}
