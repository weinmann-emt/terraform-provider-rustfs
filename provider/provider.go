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
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "MinIO server endpoint in the format host:port",
			},
			"access_key": schema.StringAttribute{
				Required: true,
			},
			"access_secret": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				Optional: true,
			},
			"ssl": schema.BoolAttribute{
				Optional: true,
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

	// Example client configuration for data sources and resources
	tr, err := minio.DefaultTransport(config.Ssl.ValueBool())
	if config.Insecure.ValueBool() {
		tr.TLSClientConfig.InsecureSkipVerify = true
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	minico_client, err := minio.New(config.Endpoint.ValueString(), &minio.Options{
		Secure:    config.Ssl.ValueBool(),
		Creds:     credentials.NewStaticV4(config.AccessKey.ValueString(), config.AccessSecret.ValueString(), ""),
		Transport: tr,
	})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	client := &AllClient{
		Minio:      minico_client,
		RustClient: rustfs.New(generatedConfig),
	}

	// if err != nil {
	// 	resp.Diagnostics.AddError(err.Error(), err.Error())
	// 	return
	// }
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *RustfsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserRessource,
		NewPolicyRessource,
		NewServiceAccountRessource,
	}
}

func (p *RustfsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RustfsProvider{
			version: version,
		}
	}
}
