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
	Id                             types.String                         `tfsdk:"id"`
	Status                         types.String                         `tfsdk:"status"`
	Filter                         *filterModel                         `tfsdk:"filter"`
	Expiration                     *expirationModel                     `tfsdk:"expiration"`
	Transition                     *transitionModel                     `tfsdk:"transition"`
	NoncurrentVersionExpiration    *noncurrentVersionExpirationModel    `tfsdk:"noncurrent_version_expiration"`
	NoncurrentVersionTransition    *noncurrentVersionTransitionModel    `tfsdk:"noncurrent_version_transition"`
	AbortIncompleteMultipartUpload *abortIncompleteMultipartUploadModel `tfsdk:"abort_incomplete_multipart_upload"`
}

type filterModel struct {
	Prefix types.String `tfsdk:"prefix"`
}

type expirationModel struct {
	Days                      types.Int64  `tfsdk:"days"`
	Date                      types.String `tfsdk:"date"`
	ExpiredObjectDeleteMarker types.Bool   `tfsdk:"expired_object_delete_marker"`
}

type transitionModel struct {
	Days         types.Int64  `tfsdk:"days"`
	Date         types.String `tfsdk:"date"`
	StorageClass types.String `tfsdk:"storage_class"`
}

type noncurrentVersionExpirationModel struct {
	NoncurrentDays types.Int64 `tfsdk:"noncurrent_days"`
}

type noncurrentVersionTransitionModel struct {
	NoncurrentDays types.Int64  `tfsdk:"noncurrent_days"`
	StorageClass   types.String `tfsdk:"storage_class"`
}

type abortIncompleteMultipartUploadModel struct {
	DaysAfterInitiation types.Int64 `tfsdk:"days_after_initiation"`
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
								"date": schema.StringAttribute{
									Optional:    true,
									Description: "Date at which the objects expire (ISO8601, e.g. 2026-12-31T00:00:00Z)",
								},
								"expired_object_delete_marker": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether to remove the delete marker of expired objects with no versions",
								},
							},
						},
						"transition": schema.SingleNestedBlock{
							Description: "Configuration block for transitioning objects to an ILM tier",
							Attributes: map[string]schema.Attribute{
								"days": schema.Int64Attribute{
									Optional:    true,
									Description: "Lifetime of the objects in days before transition",
								},
								"date": schema.StringAttribute{
									Optional:    true,
									Description: "Date at which the objects are transitioned (ISO8601, e.g. 2026-12-31T00:00:00Z)",
								},
								"storage_class": schema.StringAttribute{
									Optional:    true,
									Description: "Name of the RustFS ILM tier to transition objects to",
								},
							},
						},
						"noncurrent_version_expiration": schema.SingleNestedBlock{
							Description: "Configuration block for expiring noncurrent object versions",
							Attributes: map[string]schema.Attribute{
								"noncurrent_days": schema.Int64Attribute{
									Optional:    true,
									Description: "Number of days an object is noncurrent before it expires",
								},
							},
						},
						"noncurrent_version_transition": schema.SingleNestedBlock{
							Description: "Configuration block for transitioning noncurrent object versions to an ILM tier",
							Attributes: map[string]schema.Attribute{
								"noncurrent_days": schema.Int64Attribute{
									Optional:    true,
									Description: "Number of days an object is noncurrent before it is transitioned",
								},
								"storage_class": schema.StringAttribute{
									Optional:    true,
									Description: "Name of the RustFS ILM tier to transition noncurrent versions to",
								},
							},
						},
						"abort_incomplete_multipart_upload": schema.SingleNestedBlock{
							Description: "Configuration block for aborting incomplete multipart uploads",
							Attributes: map[string]schema.Attribute{
								"days_after_initiation": schema.Int64Attribute{
									Optional:    true,
									Description: "Number of days after multipart upload initiation before the upload is aborted",
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
			exp := &rustfs.LifecycleExpiration{}
			if !rulePlan.Expiration.Days.IsNull() {
				daysVal := int(rulePlan.Expiration.Days.ValueInt64())
				exp.Days = &daysVal
			}
			if !rulePlan.Expiration.Date.IsNull() {
				exp.Date = rulePlan.Expiration.Date.ValueString()
			}
			if !rulePlan.Expiration.ExpiredObjectDeleteMarker.IsNull() {
				marker := rulePlan.Expiration.ExpiredObjectDeleteMarker.ValueBool()
				exp.ExpiredObjectDeleteMarker = &marker
			}
			rule.Expiration = exp
		}

		if rulePlan.Transition != nil {
			tr := &rustfs.LifecycleTransition{StorageClass: rulePlan.Transition.StorageClass.ValueString()}
			if !rulePlan.Transition.Days.IsNull() {
				daysVal := int(rulePlan.Transition.Days.ValueInt64())
				tr.Days = &daysVal
			}
			if !rulePlan.Transition.Date.IsNull() {
				tr.Date = rulePlan.Transition.Date.ValueString()
			}
			rule.Transition = tr
		}

		if rulePlan.NoncurrentVersionExpiration != nil {
			ncExp := &rustfs.LifecycleNoncurrentVersionExpiration{}
			if !rulePlan.NoncurrentVersionExpiration.NoncurrentDays.IsNull() {
				daysVal := int(rulePlan.NoncurrentVersionExpiration.NoncurrentDays.ValueInt64())
				ncExp.NoncurrentDays = &daysVal
			}
			rule.NoncurrentVersionExpiration = ncExp
		}

		if rulePlan.NoncurrentVersionTransition != nil {
			ncTr := &rustfs.LifecycleNoncurrentVersionTransition{StorageClass: rulePlan.NoncurrentVersionTransition.StorageClass.ValueString()}
			if !rulePlan.NoncurrentVersionTransition.NoncurrentDays.IsNull() {
				daysVal := int(rulePlan.NoncurrentVersionTransition.NoncurrentDays.ValueInt64())
				ncTr.NoncurrentDays = &daysVal
			}
			rule.NoncurrentVersionTransition = ncTr
		}

		if rulePlan.AbortIncompleteMultipartUpload != nil {
			abort := &rustfs.LifecycleAbortIncompleteMultipartUpload{}
			if !rulePlan.AbortIncompleteMultipartUpload.DaysAfterInitiation.IsNull() {
				daysVal := int(rulePlan.AbortIncompleteMultipartUpload.DaysAfterInitiation.ValueInt64())
				abort.DaysAfterInitiation = &daysVal
			}
			rule.AbortIncompleteMultipartUpload = abort
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

		if ruleAPI.Expiration != nil {
			exp := &expirationModel{}
			if ruleAPI.Expiration.Days != nil {
				exp.Days = types.Int64Value(int64(*ruleAPI.Expiration.Days))
			}
			if ruleAPI.Expiration.Date != "" {
				exp.Date = types.StringValue(ruleAPI.Expiration.Date)
			}
			if ruleAPI.Expiration.ExpiredObjectDeleteMarker != nil {
				exp.ExpiredObjectDeleteMarker = types.BoolValue(*ruleAPI.Expiration.ExpiredObjectDeleteMarker)
			}
			rm.Expiration = exp
		}

		if ruleAPI.Transition != nil {
			tr := &transitionModel{StorageClass: types.StringValue(ruleAPI.Transition.StorageClass)}
			if ruleAPI.Transition.Days != nil {
				tr.Days = types.Int64Value(int64(*ruleAPI.Transition.Days))
			}
			if ruleAPI.Transition.Date != "" {
				tr.Date = types.StringValue(ruleAPI.Transition.Date)
			}
			rm.Transition = tr
		}

		if ruleAPI.NoncurrentVersionExpiration != nil {
			ncExp := &noncurrentVersionExpirationModel{}
			if ruleAPI.NoncurrentVersionExpiration.NoncurrentDays != nil {
				ncExp.NoncurrentDays = types.Int64Value(int64(*ruleAPI.NoncurrentVersionExpiration.NoncurrentDays))
			}
			rm.NoncurrentVersionExpiration = ncExp
		}

		if ruleAPI.NoncurrentVersionTransition != nil {
			ncTr := &noncurrentVersionTransitionModel{StorageClass: types.StringValue(ruleAPI.NoncurrentVersionTransition.StorageClass)}
			if ruleAPI.NoncurrentVersionTransition.NoncurrentDays != nil {
				ncTr.NoncurrentDays = types.Int64Value(int64(*ruleAPI.NoncurrentVersionTransition.NoncurrentDays))
			}
			rm.NoncurrentVersionTransition = ncTr
		}

		if ruleAPI.AbortIncompleteMultipartUpload != nil {
			abort := &abortIncompleteMultipartUploadModel{}
			if ruleAPI.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
				abort.DaysAfterInitiation = types.Int64Value(int64(*ruleAPI.AbortIncompleteMultipartUpload.DaysAfterInitiation))
			}
			rm.AbortIncompleteMultipartUpload = abort
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
			exp := &rustfs.LifecycleExpiration{}
			if !rulePlan.Expiration.Days.IsNull() {
				daysVal := int(*rulePlan.Expiration.Days.ValueInt64Pointer())
				exp.Days = &daysVal
			}
			if !rulePlan.Expiration.Date.IsNull() {
				exp.Date = rulePlan.Expiration.Date.ValueString()
			}
			if !rulePlan.Expiration.ExpiredObjectDeleteMarker.IsNull() {
				marker := rulePlan.Expiration.ExpiredObjectDeleteMarker.ValueBool()
				exp.ExpiredObjectDeleteMarker = &marker
			}
			rule.Expiration = exp
		}

		if rulePlan.Transition != nil {
			tr := &rustfs.LifecycleTransition{StorageClass: rulePlan.Transition.StorageClass.ValueString()}
			if !rulePlan.Transition.Days.IsNull() {
				daysVal := int(*rulePlan.Transition.Days.ValueInt64Pointer())
				tr.Days = &daysVal
			}
			if !rulePlan.Transition.Date.IsNull() {
				tr.Date = rulePlan.Transition.Date.ValueString()
			}
			rule.Transition = tr
		}

		if rulePlan.NoncurrentVersionExpiration != nil {
			ncExp := &rustfs.LifecycleNoncurrentVersionExpiration{}
			if !rulePlan.NoncurrentVersionExpiration.NoncurrentDays.IsNull() {
				daysVal := int(*rulePlan.NoncurrentVersionExpiration.NoncurrentDays.ValueInt64Pointer())
				ncExp.NoncurrentDays = &daysVal
			}
			rule.NoncurrentVersionExpiration = ncExp
		}

		if rulePlan.NoncurrentVersionTransition != nil {
			ncTr := &rustfs.LifecycleNoncurrentVersionTransition{StorageClass: rulePlan.NoncurrentVersionTransition.StorageClass.ValueString()}
			if !rulePlan.NoncurrentVersionTransition.NoncurrentDays.IsNull() {
				daysVal := int(*rulePlan.NoncurrentVersionTransition.NoncurrentDays.ValueInt64Pointer())
				ncTr.NoncurrentDays = &daysVal
			}
			rule.NoncurrentVersionTransition = ncTr
		}

		if rulePlan.AbortIncompleteMultipartUpload != nil {
			abort := &rustfs.LifecycleAbortIncompleteMultipartUpload{}
			if !rulePlan.AbortIncompleteMultipartUpload.DaysAfterInitiation.IsNull() {
				daysVal := int(*rulePlan.AbortIncompleteMultipartUpload.DaysAfterInitiation.ValueInt64Pointer())
				abort.DaysAfterInitiation = &daysVal
			}
			rule.AbortIncompleteMultipartUpload = abort
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
