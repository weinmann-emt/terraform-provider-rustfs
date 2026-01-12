package rustfs_test

import (
	"os"
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestIsAdmin(t *testing.T) {
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
	admin, _ := dut.IsAdmin()
	if !admin {
		t.Error("User is no admin")
	}

}
