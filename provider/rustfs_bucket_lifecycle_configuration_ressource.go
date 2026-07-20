package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &bucketLifecycleConfigurationRessource{}
)

// NewBucketLifecycleConfigurationRessource is a helper function to simplify the provider implementation.
func NewBucketLifecycleConfigurationRessource() resource.Resource {
	return &bucketLifecycleConfigurationRessource{}
}

// bucketLifecycleConfigurationRessource is the resource implementation.
type bucketLifecycleConfigurationRessource struct {
	client *AllClient
}

type bucketLifecycleConfigurationModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Id     types.String `tfsdk:"id"`
	Rule   []ruleModel  `tfsdk:"rule"`
}

type ruleModel struct {
	Id         types.String     `tfsdk:"id"`
	Status     types.String     `tfsdk:"status"`
	Filter     *filterModel     `tfsdk:"filter"`
	Expiration *expirationModel `tfsdk:"expiration"`
}

type filterModel struct {
	Prefix types.String `tfsdk:"prefix"`
}

type expirationModel struct {
	Days types.Int64 `tfsdk:"days"`
}

// Metadata returns the resource type name.
func (r *bucketLifecycleConfigurationRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_lifecycle_configuration"
}

// Schema defines the schema for the resource.
func (r *bucketLifecycleConfigurationRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage S3 bucket lifecycle configurations in rustfs",
		MarkdownDescription: "Manage S3 bucket lifecycle configurations in rustfs",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Name of the bucket",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The bucket name",
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				Description: "List of lifecycle rules",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "Unique identifier for the rule",
						},
						"status": schema.StringAttribute{
							Required:    true,
							Description: "Status of the rule, either Enabled or Disabled",
							Validators: []validator.String{
								stringvalidator.OneOf("Enabled", "Disabled"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"filter": schema.SingleNestedBlock{
							Description: "Filter identifying one or more objects to which the rule applies",
							Attributes: map[string]schema.Attribute{
								"prefix": schema.StringAttribute{
									Optional:    true,
									Description: "Object key prefix identifying one or more objects to which the rule applies",
								},
							},
						},
						"expiration": schema.SingleNestedBlock{
							Description: "Configuration block for object expiration",
							Attributes: map[string]schema.Attribute{
								"days": schema.Int64Attribute{
									Optional:    true,
									Description: "Lifetime of the objects in days",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *bucketLifecycleConfigurationRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *bucketLifecycleConfigurationRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketLifecycleConfigurationModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rules []rustfs.LifecycleRule
	for _, rulePlan := range plan.Rule {
		rule := rustfs.LifecycleRule{
			ID:     rulePlan.Id.ValueString(),
			Status: rulePlan.Status.ValueString(),
		}

		if rulePlan.Filter != nil {
			rule.Filter = rustfs.LifecycleFilter{
				Prefix: rulePlan.Filter.Prefix.ValueString(),
			}
		}

		if rulePlan.Expiration != nil {
			daysVal := int(rulePlan.Expiration.Days.ValueInt64())
			rule.Expiration = &rustfs.LifecycleExpiration{
				Days: &daysVal,
			}
		}

		rules = append(rules, rule)
	}

	config := &rustfs.LifecycleConfiguration{
		Rules: rules,
	}

	err := r.client.RustClient.SetBucketLifecycleConfiguration(plan.Bucket.ValueString(), config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket lifecycle configuration",
			"Could not create lifecycle configuration: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "created a bucket lifecycle configuration resource")

	plan.Id = types.StringValue(plan.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *bucketLifecycleConfigurationRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketLifecycleConfigurationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.RustClient.GetBucketLifecycleConfiguration(state.Bucket.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchLifecycleConfiguration") ||
			strings.Contains(err.Error(), "NoSuchBucket") ||
			strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading bucket lifecycle configuration",
			"Could not read lifecycle configuration: "+err.Error(),
		)
		return
	}

	state.Rule = []ruleModel{}
	for _, ruleAPI := range config.Rules {
		rm := ruleModel{
			Id:     types.StringValue(ruleAPI.ID),
			Status: types.StringValue(ruleAPI.Status),
		}

		if ruleAPI.Filter.Prefix != "" {
			rm.Filter = &filterModel{
				Prefix: types.StringValue(ruleAPI.Filter.Prefix),
			}
		}

		if ruleAPI.Expiration != nil && *ruleAPI.Expiration.Days != 0 {
			rm.Expiration = &expirationModel{
				Days: types.Int64Value(int64(*ruleAPI.Expiration.Days)),
			}
		}

		state.Rule = append(state.Rule, rm)
	}

	state.Id = types.StringValue(state.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *bucketLifecycleConfigurationRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketLifecycleConfigurationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rules []rustfs.LifecycleRule
	for _, rulePlan := range plan.Rule {
		rule := rustfs.LifecycleRule{
			ID:     rulePlan.Id.ValueString(),
			Status: rulePlan.Status.ValueString(),
		}

		if rulePlan.Filter != nil {
			rule.Filter = rustfs.LifecycleFilter{
				Prefix: rulePlan.Filter.Prefix.ValueString(),
			}
		}

		if rulePlan.Expiration != nil {
			daysVal := int(*rulePlan.Expiration.Days.ValueInt64Pointer())
			rule.Expiration = &rustfs.LifecycleExpiration{
				Days: &daysVal,
			}
		}

		rules = append(rules, rule)
	}

	config := &rustfs.LifecycleConfiguration{
		Rules: rules,
	}

	err := r.client.RustClient.SetBucketLifecycleConfiguration(plan.Bucket.ValueString(), config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bucket lifecycle configuration",
			"Could not update lifecycle configuration: "+err.Error(),
		)
		return
	}

	plan.Id = types.StringValue(plan.Bucket.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *bucketLifecycleConfigurationRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketLifecycleConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RustClient.DeleteBucketLifecycleConfiguration(data.Bucket.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchLifecycleConfiguration") ||
			strings.Contains(err.Error(), "NoSuchBucket") ||
			strings.Contains(err.Error(), "404") {
			// Already deleted
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting bucket lifecycle configuration",
			"Could not delete lifecycle configuration: "+err.Error(),
		)
		return
	}
}
