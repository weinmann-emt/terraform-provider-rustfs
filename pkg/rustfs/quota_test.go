package rustfs_test

import (
	"strings"
	"testing"
	"time"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestReadQuota(t *testing.T) {
	name := randomString()
	dut := getClient()
	name = strings.ToLower(name)
	if err := dut.CreateBucket(name); err != nil {
		t.Fatal(err)
	}
	resp, err := dut.ReadQuota(name)
	if err != nil {
		t.Error(err)
	}
	if resp.Bucket != name {
		t.Error("Bucket readback unexpected value")
	}
	if err := dut.DeleteBucket(name); err != nil {
		t.Fatal(err)
	}
}

func TestCRDQuota(t *testing.T) {
	name := randomString()
	name = strings.ToLower(name)
	quota := rustfs.Quota{
		Bucket: name,
		Quota:  100054541,
	}
	dut := getClient()
	if err := dut.CreateBucket(name); err != nil {
		t.Fatal(err)
	}
	_, err := dut.ReadQuota(name)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	resp, err := dut.SetQuota(quota)
	if err != nil {
		t.Error(err)
	}
	resp, err = dut.ReadQuota(name)
	if err != nil {
		t.Error(err)
	}
	if resp.Quota != quota.Quota {
		t.Error("Readback gave wrong quota")
	}

	if err := dut.DeletQuota(name); err != nil {
		t.Error("error during quota remove")
	}
}
