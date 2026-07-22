package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RebalanceResource{}

type RebalanceResource struct {
	client *AllClient
}

type RebalanceResourceModel struct {
	ID types.String `tfsdk:"id"`
}

func NewRebalanceResource() resource.Resource {
	return &RebalanceResource{}
}

func (r *RebalanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rebalance"
}

func (r *RebalanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Trigger RustFS pool rebalancing",
		MarkdownDescription: "Triggers a pool rebalancing operation in RustFS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Rebalance trigger ID.",
			},
		},
	}
}

func (r *RebalanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RebalanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RebalanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RustClient.StartRebalance(); err != nil {
		resp.Diagnostics.AddError("Error starting rebalance", "Could not start rebalance: "+err.Error())
		return
	}

	plan.ID = types.StringValue("rebalance-triggered")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RebalanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RebalanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RebalanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &req.Plan)...)
}

func (r *RebalanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
