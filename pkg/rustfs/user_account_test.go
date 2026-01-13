package rustfs_test

import (
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestCreateUserAccount(t *testing.T) {

	account := rustfs.UserAccount{
		AccessKey: randomString(8),
		SecretKey: randomString(8),
	}
	dut := getClient()
	err := dut.CreateUserAccount(account)
	if err != nil {
		t.Error(err)
	}
}
