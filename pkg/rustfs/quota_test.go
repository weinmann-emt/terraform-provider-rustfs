package rustfs_test

import (
	"strings"
	"testing"
	"time"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)


func TestReadQuota(t *testing.T) {
	name :=  randomString(8)
	dut := getClient()
	name = strings.ToLower(name)
	dut.CreateBucket(name)
	resp, err := dut.ReadQuota(name)
	if err != nil {
		t.Error(err)
	}
	if resp.Bucket != name {
		t.Error("Bucket readback unexpected value")
	}
	dut.DeleteBucket(name)
}

func TestCRDQuota(t *testing.T){
	name :=  randomString(8)
	name = strings.ToLower(name)
	quota := rustfs.Quota{
		Bucket: name,
		Quota: 100054541,
	}
	dut := getClient()
	dut.CreateBucket(name)
	resp, err := dut.ReadQuota(name)
	if resp.Bucket != name {
		t.Error("Bucket readback unexpected value")
	}
	time.Sleep(5)
	resp, err = dut.SetQuota(quota)
	if err != nil {
		t.Error(err)
	}
	resp, err = dut.ReadQuota(name)
	if resp.Quota != quota.Quota {
		t.Error("Readback gave wring quota")
	}

	err = dut.DeletQuota(name)
	if err != nil {
		t.Error("error during quota remove")
	}


}
