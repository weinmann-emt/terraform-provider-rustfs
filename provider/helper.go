package provider

import (
	"os"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func generateRustClientConfig(model RustfsProviderModel) *rustfs.RustfsAdminConfig {

	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	if endpoint == "" {
		endpoint = model.Endpoint.ValueString()
	}

	user := os.Getenv("RUSTFS_USER")
	if user == "" {
		user = model.AccessKey.ValueString()
	}

	secret := os.Getenv("RUSTFS_SECRET")
	if secret == "" {
		secret = model.AccessSecret.ValueString()
	}

	config := &rustfs.RustfsAdminConfig{
		Endpoint:     endpoint,
		AccessKey:    user,
		AccessSecret: secret,
		Ssl:          model.Ssl.ValueBool(),
		Insecure:     model.Insecure.ValueBool(),
	}
	return config
}
