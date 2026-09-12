package rustfs_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

// newTestAdminServer spins up an httptest server speaking the RustFS admin
// API and returns a RustfsAdmin client pointed at it.
func newTestAdminServer(t *testing.T, handler http.HandlerFunc) *rustfs.RustfsAdmin {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := rustfs.New(&rustfs.RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(srv.URL, "http://"),
	})
	return &c
}

func TestIsAdmin(t *testing.T) {
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
	admin, _ := dut.IsAdmin()
	if !admin {
		t.Error("User is no admin")
	}

}
