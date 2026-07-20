package rustfs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		pools := []PoolInfo{
			{Name: "pool-0"},
			{Name: "pool-1"},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pools)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	pools, err := client.ListPools()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(pools))
	}
	if pools[0].Name != "pool-0" {
		t.Errorf("expected pool-0, got %s", pools[0].Name)
	}
}
