package rustfs_test

import (
	"strings"
	"testing"
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


