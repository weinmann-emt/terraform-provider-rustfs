package rustfs_test

import (
	"os"
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestIsAdmin(t *testing.T) {
	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	key := os.Getenv("RUSTFS_KEY")
	secret := os.Getenv("RUSTFS_SECRET")

	config := rustfs.RustfsAdminConfig{
		AccessKey:    key,
		AccessSecret: secret,
		Endpoint:     endpoint,

		Ssl: false,
	}

	dut, _ := rustfs.New(config)
	admin, _ := dut.IsAdmin()
	if !admin {
		t.Error("User is no admin")
	}

}
