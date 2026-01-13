package rustfs_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func getClient() rustfs.RustfsAdmin {
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
	return dut
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

func TestCreateServiceAccount(t *testing.T) {

	account := rustfs.ServiceAccount{
		AccessKey: randomString(8),
		SecretKey: "someSuperS3cret",
		Name:      randomString(8),
	}
	dut := getClient()
	err := dut.CreateServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateAndDeleteServiceAccount(t *testing.T) {

	account := rustfs.ServiceAccount{
		AccessKey: randomString(8),
		SecretKey: "someSuperS3cret",
		Name:      randomString(8),
	}
	dut := getClient()
	err := dut.CreateServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
	err = dut.DeleteServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
}
func TestCreateUpdateAndDeleteServiceAccount(t *testing.T) {

	account := rustfs.ServiceAccount{
		AccessKey: randomString(8),
		SecretKey: "someSuperS3cret",
		Name:      randomString(8),
	}
	dut := getClient()
	err := dut.CreateServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
	account.SecretKey = "insecureOne"
	err = dut.UpdateServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
	err = dut.DeleteServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
}
