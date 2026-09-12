package rustfs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTierStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rustfs/admin/v3/tier-stats") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"WARM":{"total_size":15,"num_versions":3,"num_objects":1},"ARCHIVE":{"total_size":9,"num_versions":1,"num_objects":1}}`))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	stats, err := client.TierStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 tiers, got %d: %v", len(stats), stats)
	}
	warm, ok := stats["WARM"]
	if !ok {
		t.Fatalf("expected WARM tier, got %v", stats)
	}
	if warm.TotalSize != 15 || warm.NumVersions != 3 || warm.NumObjects != 1 {
		t.Errorf("unexpected WARM stats: %+v", warm)
	}
	archive, ok := stats["ARCHIVE"]
	if !ok {
		t.Fatalf("expected ARCHIVE tier, got %v", stats)
	}
	if archive.TotalSize != 9 || archive.NumVersions != 1 || archive.NumObjects != 1 {
		t.Errorf("unexpected ARCHIVE stats: %+v", archive)
	}
}

func TestTierStatsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	stats, err := client.TierStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected no tiers, got %v", stats)
	}
}
