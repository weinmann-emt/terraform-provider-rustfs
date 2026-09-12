package provider

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/minio/minio-go/v7/pkg/signer"
)

func TestAccGroupsDataSource(t *testing.T) {
	groupName := fmt.Sprintf("tf-acc-groups-%d", time.Now().UnixNano())
	createAccTestGroup(t, groupName)
	defer deleteAccTestGroup(t, groupName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "rustfs_groups" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.rustfs_groups.all", "groups.#"),
					resource.TestCheckTypeSetElemAttr("data.rustfs_groups.all", "groups.*", groupName),
				),
			},
		},
	})
}

func createAccTestGroup(t *testing.T, name string) {
	t.Helper()
	body := []byte(fmt.Sprintf(`{"group":%q,"members":[],"isRemove":false,"groupStatus":"enabled"}`, name))
	accAdminRequest(t, http.MethodPut, "update-group-members", body)
}

func deleteAccTestGroup(t *testing.T, name string) {
	t.Helper()
	accAdminRequest(t, http.MethodDelete, "group/"+name, nil)
}

func accAdminRequest(t *testing.T, method, relPath string, body []byte) {
	t.Helper()
	url := "http://" + os.Getenv("RUSTFS_ENDPOINT") + "/rustfs/admin/v3/" + relPath
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	if len(body) > 0 {
		req.ContentLength = int64(len(body))
	}
	sum := sha256.Sum256(body)
	req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sum[:]))
	req = signer.SignV4(*req, os.Getenv("RUSTFS_USER"), os.Getenv("RUSTFS_SECRET"), "", "us-east-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("admin request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		t.Fatalf("admin request %s %s returned status %d", method, relPath, resp.StatusCode)
	}
}
