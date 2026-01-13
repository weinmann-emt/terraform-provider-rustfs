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
	err = dut.DeleteUserAccount(account)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateUserAccountWithGrp(t *testing.T) {

	account := rustfs.UserAccount{
		AccessKey: randomString(8),
		SecretKey: randomString(8),
		Group:     "readwrite",
	}
	dut := getClient()
	err := dut.CreateUserAccount(account)
	if err != nil {
		t.Error(err)
	}
	err = dut.DeleteUserAccount(account)
	if err != nil {
		t.Error(err)
	}
}

func TestAddUserWithAccesKey(t *testing.T) {
	account := rustfs.UserAccount{
		AccessKey: randomString(8),
		SecretKey: randomString(8),
	}
	dut := getClient()
	err := dut.CreateUserAccount(account)
	if err != nil {
		t.Error(err)
	}

	service := rustfs.ServiceAccount{
		AccessKey:  randomString(8),
		SecretKey:  "someSuperS3cret",
		Name:       randomString(8),
		TargetUser: account.AccessKey,
	}
	err = dut.CreateServiceAccount(service)
	if err != nil {
		t.Error(err)
	}

	err = dut.DeleteServiceAccount(service)
	if err != nil {
		t.Error(err)
	}
	err = dut.DeleteUserAccount(account)
	if err != nil {
		t.Error(err)
	}

}
