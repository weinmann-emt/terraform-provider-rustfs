package rustfs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/info" {
			t.Errorf("expected path /rustfs/admin/v3/info, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(sampleServerInfoPayload()); err != nil {
			t.Errorf("error encoding payload: %v", err)
		}
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	info, err := client.ServerInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Info.Mode != "online" {
		t.Errorf("expected mode online, got %q", info.Info.Mode)
	}
	if info.Info.DeploymentID != "3d6a83f0-8b13-49e0-bd0e-3a921f256a6c" {
		t.Errorf("unexpected deployment id %q", info.Info.DeploymentID)
	}
	if len(info.Info.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(info.Info.Servers))
	}
	serverEntry := info.Info.Servers[0]
	if serverEntry.Version != "2026-09-01T17:35:49Z@1.0.0-rc.5" {
		t.Errorf("unexpected server version %q", serverEntry.Version)
	}
	if serverEntry.Uptime != 18986 {
		t.Errorf("expected uptime 18986, got %d", serverEntry.Uptime)
	}
	if len(serverEntry.Drives) != 1 {
		t.Fatalf("expected 1 drive, got %d", len(serverEntry.Drives))
	}
	if serverEntry.Drives[0].State != "ok" {
		t.Errorf("expected drive state ok, got %q", serverEntry.Drives[0].State)
	}
	if info.Info.Backend.BackendType != "Erasure" {
		t.Errorf("expected backend type Erasure, got %q", info.Info.Backend.BackendType)
	}
	if info.Info.Buckets.Count != 5 {
		t.Errorf("expected 5 buckets, got %d", info.Info.Buckets.Count)
	}
	if info.Info.Usage.Size != 0 {
		t.Errorf("expected usage size 0, got %d", info.Info.Usage.Size)
	}
}

func sampleServerInfoPayload() map[string]interface{} {
	return map[string]interface{}{
		"admin_discovery": map[string]interface{}{
			"clusterSnapshot":     "/rustfs/admin/v4/cluster/snapshot",
			"extensionsCatalog":   "/rustfs/admin/v4/extensions/catalog",
			"runtimeCapabilities": "/rustfs/admin/v4/runtime/capabilities",
		},
		"bitrotSelftest": "passed",
		"info": map[string]interface{}{
			"backend": map[string]interface{}{
				"backendType":       "Erasure",
				"offlineDisks":      0,
				"onlineDisks":       1,
				"rrSCParity":        0,
				"standardSCParity":  0,
				"totalDrivesPerSet": []interface{}{1},
				"totalSets":         []interface{}{1},
				"unknownDisks":      0,
			},
			"buckets": map[string]interface{}{
				"count": 5,
				"error": nil,
			},
			"deletemarkers": map[string]interface{}{
				"count": 0,
				"error": nil,
			},
			"deploymentID": "3d6a83f0-8b13-49e0-bd0e-3a921f256a6c",
			"domain":       nil,
			"mode":         "online",
			"objects": map[string]interface{}{
				"count": 0,
				"error": nil,
			},
			"pools": map[string]interface{}{
				"0": map[string]interface{}{
					"0": map[string]interface{}{
						"deleteMarkersCount": 0,
						"healDisks":          0,
						"id":                 0,
						"objectsCount":       0,
						"rawCapacity":        63728975872,
						"rawUsage":           4699373568,
						"usage":              0,
						"versionsCount":      0,
					},
				},
			},
			"region": nil,
			"servers": []interface{}{
				map[string]interface{}{
					"commitID": "",
					"drives": []interface{}{
						map[string]interface{}{
							"availspace":   59029602304,
							"endpoint":     "/data",
							"healing":      false,
							"local":        true,
							"path":         "/data",
							"runtimeState": "online",
							"state":        "ok",
							"totalspace":   63728975872,
							"usedspace":    4699373568,
							"uuid":         "14b20c88-71bd-4d67-aaf3-bf2446b1b153",
							"utilization":  7.373998253853503,
						},
					},
					"endpoint":  ":::9000",
					"max_procs": 6,
					"mem_stats": map[string]interface{}{
						"alloc":       180760576,
						"frees":       0,
						"heap_alloc":  180760576,
						"mallocs":     0,
						"total_alloc": 841904128,
					},
					"num_cpu": 6,
					"state":   "online",
					"uptime":  18986,
					"version": "2026-09-01T17:35:49Z@1.0.0-rc.5",
				},
			},
			"usage": map[string]interface{}{
				"error": nil,
				"size":  0,
			},
			"versions": map[string]interface{}{
				"count": 0,
				"error": nil,
			},
		},
	}
}
