// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// Ensure RustfsProvider satisfies various provider interfaces.
var _ provider.Provider = &RustfsProvider{}

// RustfsProvider defines the provider implementation.
type RustfsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// RustfsProviderModel describes the provider data model.
type RustfsProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	AccessKey    types.String `tfsdk:"access_key"`
	AccessSecret types.String `tfsdk:"access_secret"`
	Ssl          types.Bool   `tfsdk:"ssl"`
	Insecure     types.Bool   `tfsdk:"insecure"`
}

func (p *RustfsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rustfs"
	resp.Version = p.version
}

func (p *RustfsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Interact with rustfs",
		MarkdownDescription: "Provider to access with RustFS",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "RUSTFS server endpoint in the format host:port. Defaults to RUSTFS_ENDPOINT environment variable.",
			},
			"access_key": schema.StringAttribute{
				Optional:    true,
				Description: "Username or access key. Defaults to RUSTFS_USER environment variable.",
			},
			"access_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Secret to be used as pass. Defaults to RUSTFS_SECRET environment variable.",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Insecure skip SSL validation",
			},
			"ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "Use SSL transport",
			},
		},
	}
}

func (p *RustfsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config RustfsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	generatedConfig := generateRustClientConfig(config)

	endpoint := envOrDefault("RUSTFS_ENDPOINT", config.Endpoint.ValueString())
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing RUSTFS endpoint",
			"Set the endpoint in the provider block or via the RUSTFS_ENDPOINT environment variable.",
		)
		return
	}

	accessKey := envOrDefault("RUSTFS_USER", config.AccessKey.ValueString())
	secretKey := envOrDefault("RUSTFS_SECRET", config.AccessSecret.ValueString())

	// Example client configuration for data sources and resources
	tr, err := minio.DefaultTransport(config.Ssl.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	usEast01 := "us-east-1"
	minio_client, err := minio.New(endpoint, &minio.Options{
		Secure:    config.Ssl.ValueBool(),
		Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
		Transport: tr,
		Region:    usEast01,
	})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	client := &AllClient{
		Minio:      minio_client,
		RustClient: rustfs.New(generatedConfig),
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *RustfsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserRessource,
		NewPolicyRessource,
		NewServiceAccountRessource,
		NewBucketRessource,
		NewquotaRessource,
		NewGroupResource,
		NewBucketLifecycleConfigurationRessource,
		NewTierResource,
		NewBucketObjectLockResource,
		NewBucketReplicationResource,
		NewBucketEncryptionResource,
		NewBucketVersioningResource,
	}
}

func (p *RustfsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPoolsDataSource,
		NewUsersDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RustfsProvider{
			version: version,
		}
	}
}
