package rustfs_test

import (
	"strings"
	"testing"
	"time"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestCreateUpdateDelete(t *testing.T) {
	name := randomString(8)
	name = strings.ToLower(name)
	days := 20
	dut := getClient()
	dut.CreateBucket(name)

	lifecycleConfig := rustfs.LifecycleConfiguration{
		Rules: []rustfs.LifecycleRule{
			{
				ID:     "TestRule",
				Status: "Enabled",
				Filter: rustfs.LifecycleFilter{
					Prefix: "test",
				},
				Expiration: &rustfs.LifecycleExpiration{
					Days: &days,
				},
			},
		},
	}

	err := dut.SetBucketLifecycleConfiguration(name, &lifecycleConfig)
	if err != nil {
		t.Error("Eror during create", err)
	}
	time.Sleep(5 * time.Second)

	lifecycleConfig.Rules[0].Filter.Prefix = ""
	lifecycleConfig.Rules = append(lifecycleConfig.Rules, rustfs.LifecycleRule{
		ID:     "TestRule2",
		Status: "Disabled",
		Filter: rustfs.LifecycleFilter{
			Prefix: "test",
		},
		Expiration: &rustfs.LifecycleExpiration{
			Days: &days,
		},
	})
	err = dut.SetBucketLifecycleConfiguration(name, &lifecycleConfig)
	if err != nil {
		t.Error("Eror during update", err)
	}
	time.Sleep(5 * time.Second)

	err = dut.DeleteBucketLifecycleConfiguration(name)
	if err != nil {
		t.Error("Eror during delete", err)
	}

	dut.DeleteBucket(name)
}
