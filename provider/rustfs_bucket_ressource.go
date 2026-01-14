// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/minio/minio-go/v7"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &bucketRessource{}
)

// NewbucketRessource is a helper function to simplify the provider implementation.
func NewBucketRessource() resource.Resource {
	return &bucketRessource{}
}

// bucketRessource is the resource implementation.
type bucketRessource struct {
	client *AllClient
}

type bucketResourceModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the resource type name.
func (r *bucketRessource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

// Schema defines the schema for the resource.
func (r *bucketRessource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage S3 buckets in rustfs",
		MarkdownDescription: "Manage S3 buckets in rustfs",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the bucket",
			},
		},
	}
}

func (r *bucketRessource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *bucketRessource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketResourceModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	exists, err := r.client.Minio.BucketExists(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}
	if exists {
		resp.Diagnostics.AddError(
			"Error creating bucket",
			"Already existing",
		)
		return
	}

	err = r.client.Minio.MakeBucket(ctx, plan.Name.ValueString(), minio.MakeBucketOptions{
		Region: "us-east-01",
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bucket",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, "created a resource")
	// plan.ID = types.StringValue(account.AccessKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *bucketRessource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Save update status
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *bucketRessource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *bucketRessource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Minio.RemoveBucket(ctx, data.Name.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
	}
}
