package rustfs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleStorageInfoPayload = `{
  "admin_discovery": {
    "clusterSnapshot": "/rustfs/admin/v4/cluster/snapshot",
    "extensionsCatalog": "/rustfs/admin/v4/extensions/catalog",
    "runtimeCapabilities": "/rustfs/admin/v4/runtime/capabilities"
  },
  "info": {
    "backend": {
      "BackendType": "Erasure",
      "DrivesPerSet": [1],
      "OfflineDisks": {},
      "OnlineDisks": {},
      "RRSCData": [1],
      "RRSCParities": [0],
      "RRSCParity": 0,
      "StandardSCData": [1],
      "StandardSCParities": [0],
      "StandardSCParity": 0,
      "TotalSets": [1]
    },
    "disks": [
      {
        "availspace": 59029614592,
        "disk_index": 0,
        "endpoint": "/data",
        "free_inodes": 31125287,
        "healing": false,
        "local": true,
        "major": 253,
        "minor": 4,
        "path": "/data",
        "pool_index": 0,
        "rootDisk": false,
        "runtimeState": "online",
        "scanning": false,
        "set_index": 0,
        "state": "ok",
        "totalspace": 63728975872,
        "used_inodes": 69321,
        "usedspace": 4699361280,
        "uuid": "14b20c88-71bd-4d67-aaf3-bf2446b1b153"
      }
    ]
  }
}`

func newStorageInfoTestServer(t *testing.T, body string, status int) *RustfsAdmin {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rustfs/admin/v3/storageinfo") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != "" {
			_, _ = w.Write([]byte(body))
		}
	}))
	t.Cleanup(srv.Close)

	client := New(&RustfsAdminConfig{
		Endpoint:  srv.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"
	return &client
}

func TestStorageInfo(t *testing.T) {
	client := newStorageInfoTestServer(t, sampleStorageInfoPayload, http.StatusOK)

	info, err := client.StorageInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Info == nil {
		t.Fatal("expected info details, got nil")
	}
	if info.Info.Backend == nil {
		t.Fatal("expected backend, got nil")
	}
	if info.Info.Backend.BackendType != "Erasure" {
		t.Errorf("expected BackendType Erasure, got %q", info.Info.Backend.BackendType)
	}
	if len(info.Info.Disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(info.Info.Disks))
	}
	disk := info.Info.Disks[0]
	if disk.State != "ok" {
		t.Errorf("expected disk state ok, got %q", disk.State)
	}
	if disk.TotalSpace != 63728975872 {
		t.Errorf("expected totalspace 63728975872, got %d", disk.TotalSpace)
	}
	if disk.UsedSpace != 4699361280 {
		t.Errorf("expected usedspace 4699361280, got %d", disk.UsedSpace)
	}
	if disk.AvailSpace != 59029614592 {
		t.Errorf("expected availspace 59029614592, got %d", disk.AvailSpace)
	}
	if disk.UUID != "14b20c88-71bd-4d67-aaf3-bf2446b1b153" {
		t.Errorf("unexpected uuid: %q", disk.UUID)
	}
}

func TestStorageInfoRejectsNonSuccess(t *testing.T) {
	client := newStorageInfoTestServer(t, `{"error":"boom"}`, http.StatusInternalServerError)

	_, err := client.StorageInfo()
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestStorageInfoValidJSONRoundTrip(t *testing.T) {
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(sampleStorageInfoPayload), &decoded); err != nil {
		t.Fatalf("sample payload is not valid JSON: %v", err)
	}
	if _, ok := decoded["info"]; !ok {
		t.Fatal("sample payload missing info field")
	}
	if _, ok := decoded["admin_discovery"]; !ok {
		t.Fatal("sample payload missing admin_discovery field")
	}
}
