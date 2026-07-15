package provider

import (
	"os"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func envOrDefault(envKey, defaultValue string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultValue
}

func generateRustClientConfig(model RustfsProviderModel) *rustfs.RustfsAdminConfig {
	config := &rustfs.RustfsAdminConfig{
		Endpoint:     envOrDefault("RUSTFS_ENDPOINT", model.Endpoint.ValueString()),
		AccessKey:    envOrDefault("RUSTFS_USER", model.AccessKey.ValueString()),
		AccessSecret: envOrDefault("RUSTFS_SECRET", model.AccessSecret.ValueString()),
		Ssl:          model.Ssl.ValueBool(),
		Insecure:     model.Insecure.ValueBool(),
	}
	return config
}
