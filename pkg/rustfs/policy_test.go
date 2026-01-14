package rustfs_test

import (
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func TestCreateAndDeletePolicy(t *testing.T) {
	dut := getClient()
	actions := [1]string{
		"s3:GetObject",
	}
	resources :=
		[1]string{
			"arn:aws:s3:::bucket/*",
		}
	name := "test"
	statements := []rustfs.PolicyStatement{
		{
			Effect:    "Allow",
			Action:    actions[:],
			Ressource: resources[:],
		},
	}
	policy := rustfs.Policy{
		Name:      name,
		Statement: statements,
	}
	err := dut.CreatePolicy(policy)
	if err != nil {
		t.Error(err)
	}

	read, _ := dut.ReadPolicy(policy.Name)
	if read.Name != policy.Name {
		t.Error("read back not working")
	}

	err = dut.DeletePolicy(name)
	if err != nil {
		t.Error(err)
	}
}
