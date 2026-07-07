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
	key := os.Getenv("RUSTFS_USER")
	secret := os.Getenv("RUSTFS_SECRET")

	config := rustfs.RustfsAdminConfig{
		AccessKey:    key,
		AccessSecret: secret,
		Endpoint:     endpoint,

		Ssl: false,
	}

	dut := rustfs.New(&config)
	return dut
}

<<<<<<< HEAD
func randomString() string {
=======
func randomString(length int) string {
>>>>>>> cc5174d (fix: check quota API errors, remove debug print, remove dead code, fix test env var and rand.Seed (#38))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
	}

	return string(result)
}

func TestCreateServiceAccount(t *testing.T) {

	account := rustfs.ServiceAccount{
		AccessKey: randomString(),
		SecretKey: "someSuperS3cret",
		Name:      randomString(),
	}
	dut := getClient()
	err := dut.CreateServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateAndDeleteServiceAccount(t *testing.T) {

	account := rustfs.ServiceAccount{
		AccessKey: randomString(),
		SecretKey: "someSuperS3cret",
		Name:      randomString(),
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
		AccessKey: randomString(),
		SecretKey: "someSuperS3cret",
		Name:      randomString(),
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
func TestCreateReadAndDeleteServiceAccount(t *testing.T) {

	account := rustfs.ServiceAccount{
		AccessKey: randomString(),
		SecretKey: "someSuperS3cret",
		Name:      randomString(),
	}
	dut := getClient()
	err := dut.CreateServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
	reply, err := dut.ReadServiceAccount(account.AccessKey)
	if err != nil {
		t.Error(err)
	}
	if reply.Name != account.Name {
		t.Error("Read value not matching")
	}
	err = dut.DeleteServiceAccount(account)
	if err != nil {
		t.Error(err)
	}
}
