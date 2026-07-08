package rustfs_test

import (
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestCreateUserAccount(t *testing.T) {

	account := rustfs.UserAccount{
		AccessKey: randomString(),
		SecretKey: randomString(),
	}
	dut := getClient()
	err := dut.CreateUserAccount(account)
	if err != nil {
		t.Error(err)
	}
	read, err := dut.ReadUserAccount(account.AccessKey)
	if err != nil {
		t.Error(err)
	}
	if read.Status != "enabled" {
		t.Error("wtf")
	}

	err = dut.DeleteUserAccount(account)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateUserAccountWithGrp(t *testing.T) {

	account := rustfs.UserAccount{
		AccessKey: randomString(),
		SecretKey: randomString(),
		Policy:    "readwrite",
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
		AccessKey: randomString(),
		SecretKey: randomString(),
	}
	dut := getClient()
	err := dut.CreateUserAccount(account)
	if err != nil {
		t.Error(err)
	}

	service := rustfs.ServiceAccount{
		AccessKey:  randomString(),
		SecretKey:  "someSuperS3cret",
		Name:       randomString(),
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
