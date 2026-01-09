package provider

import (
	"testing"
)

func TestConfigParse(t *testing.T) {
	config, _ := generateMinioConfig(RustfsProviderModel{})
	if len(config.S3HostPort) < 1 {
		t.Error("Error baking config")
	}
}
