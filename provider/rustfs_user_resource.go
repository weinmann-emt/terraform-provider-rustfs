package provider

import (
	"context"

	"github.com/aminueza/terraform-provider-minio/minio"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ExampleResource{}
var _ resource.ResourceWithImportState = &ExampleResource{}

// ExampleResource defines the resource implementation.
type RustfsUserRessource struct {
	client *minio.S3MinioClient
}

type RustfsUserRessourceModel struct {
	AccessKey types.String `tfsdk:"accessKey"`
	SecretKey types.String `tfsdk:"secretKey"`
	Status    types.String `tfsdk:"status"`
	Group     types.String `tfsdk:"group"`
	Id        types.String `tfsdk:"id"`
}

func (r *RustfsUserRessource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *RustfsUserRessource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "RustFS user",

		Attributes: map[string]schema.Attribute{
			"accessKey": schema.StringAttribute{
				MarkdownDescription: "Access Key",
				Optional:            false,
			},
			"secretKey": schema.StringAttribute{
				MarkdownDescription: "Secret Key",
				Optional:            false,
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status",
			},
			"group": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User Group",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// func (r *RustfsUserRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

// 	var data RustfsUserRessourceModel
// 	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

// 	if resp.Diagnostics.HasError() {
// 		return
// 	}
// 	data.Id = data.AccessKey

// 	account := rustfs.UserAccount{
// 		AccessKey: data.AccessKey.String(),
// 		SecretKey: data.SecretKey.String(),
// 		Group:     data.Group.String(),
// 		Status:    data.Status.String(),
// 	}

// 	tflog.Trace(ctx, "created a resource")
// 	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
// }
