package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStorageInfo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "rustfs_storage_info" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rustfs_storage_info.test", "raw_json"),
					resource.TestCheckResourceAttrSet("data.rustfs_storage_info.test", "backend.backend_type"),
					resource.TestCheckResourceAttrSet("data.rustfs_storage_info.test", "disks.#"),
					resource.TestCheckResourceAttrSet("data.rustfs_storage_info.test", "disks.0.path"),
					resource.TestCheckResourceAttrSet("data.rustfs_storage_info.test", "disks.0.total_space"),
					resource.TestCheckResourceAttrSet("data.rustfs_storage_info.test", "disks.0.state"),
				),
			},
			{
				Config: testAccProviderConfig() + `
data "rustfs_storage_info" "test" {}

output "storage_raw" {
  value = data.rustfs_storage_info.test.raw_json
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rustfs_storage_info.test", "raw_json"),
					resource.TestCheckResourceAttrWith("data.rustfs_storage_info.test", "raw_json", func(v string) error {
						var parsed map[string]interface{}
						if err := json.Unmarshal([]byte(v), &parsed); err != nil {
							return err
						}
						if _, ok := parsed["info"]; !ok {
							t.Errorf("raw_json missing info field")
						}
						return nil
					}),
				),
			},
		},
	})
}
