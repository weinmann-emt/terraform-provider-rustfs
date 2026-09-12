package rustfs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKmsStatus(t *testing.T) {
	statusJSON := `{
		"backend_type": "local",
		"backend_status": "healthy",
		"cache_enabled": true,
		"cache_stats": {
			"hit_count": 42,
			"miss_count": 7,
			"entry_count": 128,
			"eviction_count": 3
		},
		"default_key_id": "rustfs-master",
		"capabilities": {
			"encrypt": true,
			"decrypt": true,
			"generate_data_key": true,
			"rotate": false,
			"enable_disable": false,
			"schedule_deletion": false,
			"versioning": false,
			"physical_delete": false,
			"update_key_metadata": false,
			"rewrap": false,
			"production_supported": false
		},
		"cluster_config": {
			"consistent": true,
			"nodes": [
				{"host": "local", "config_fingerprint": "sha256:abc123", "error": null}
			]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/status" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statusJSON))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	status, err := client.KmsStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.BackendType != "local" {
		t.Errorf("expected backend_type local, got %s", status.BackendType)
	}
	if status.BackendStatus != "healthy" {
		t.Errorf("expected backend_status healthy, got %s", status.BackendStatus)
	}
	if !status.CacheEnabled {
		t.Error("expected cache_enabled true")
	}
	if status.CacheStats == nil {
		t.Fatal("expected cache_stats to be present")
	}
	if status.CacheStats.HitCount != 42 {
		t.Errorf("expected hit_count 42, got %d", status.CacheStats.HitCount)
	}
	if status.DefaultKeyID == nil || *status.DefaultKeyID != "rustfs-master" {
		t.Errorf("unexpected default_key_id: %v", status.DefaultKeyID)
	}
	if status.Capabilities == nil || !status.Capabilities.Encrypt {
		t.Error("expected encrypt capability true")
	}
	if status.ClusterConfig == nil || !status.ClusterConfig.Consistent {
		t.Error("expected cluster_config.consistent true")
	}
	if len(status.ClusterConfig.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(status.ClusterConfig.Nodes))
	}
}

func TestKmsConfig(t *testing.T) {
	configJSON := `{
		"backend": "vault-kv2",
		"cache_enabled": true,
		"cache_max_keys": 2048,
		"cache_ttl_seconds": 600,
		"default_key_id": "rustfs-master"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/config" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(configJSON))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	config, err := client.KmsConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Backend != "vault-kv2" {
		t.Errorf("expected backend vault-kv2, got %s", config.Backend)
	}
	if !config.CacheEnabled {
		t.Error("expected cache_enabled true")
	}
	if config.CacheMaxKeys != 2048 {
		t.Errorf("expected cache_max_keys 2048, got %d", config.CacheMaxKeys)
	}
	if config.CacheTTLSeconds != 600 {
		t.Errorf("expected cache_ttl_seconds 600, got %d", config.CacheTTLSeconds)
	}
	if config.DefaultKeyID == nil || *config.DefaultKeyID != "rustfs-master" {
		t.Errorf("unexpected default_key_id: %v", config.DefaultKeyID)
	}
}

func TestKmsStatusDecodesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"Code":    "InternalError",
			"Message": "KMS service not initialized",
		})
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	if _, err := client.KmsStatus(); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
