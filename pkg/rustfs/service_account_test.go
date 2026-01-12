package rustfs_test

import (
	"os"
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func getClient() rustfs.RustfsAdmin {
	endpoint := os.Getenv("RUSTFS_SERVER")
	key := os.Getenv("RUSTFS_USER")
	secret := os.Getenv("RUSTFS_SECRET")

	config := rustfs.RustfsAdminConfig{
		AccessKey:    key,
		AccessSecret: secret,
		Endpoint:     endpoint,

		Secure: false,
	}

	dut, _ := rustfs.New(config)
	return dut
}

func TestCreateServiceAccount(t *testing.T) {
	account := rustfs.ServiceAccount{
		AccessKey: "gocreated",
		SecretKey: "someSuperS3cret",
		Name:      "juhuay",
	}
	dut := getClient()
	err := dut.CreateServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
}
