package rustfs_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

func newTestConfigAdminServer(t *testing.T, handler http.HandlerFunc) *rustfs.RustfsAdmin {
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

func TestGetConfig(t *testing.T) {
	var gotPath string
	client := newTestConfigAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(
			"# RUSTFS_NOTIFY_WEBHOOK_ENDPOINT=http://example.com/rustfs/events\n" +
				`notify_webhook:primary endpoint="http://example.com/rustfs/events" enable="off"`,
		))
	})
	kvs, err := client.GetConfig("notify_webhook:primary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/get-config-kv") || !strings.Contains(gotPath, "key=notify_webhook%3Aprimary") {
		t.Errorf("wrong request: %s", gotPath)
	}
	if len(kvs) != 2 {
		t.Fatalf("expected 2 kvs, got %d: %+v", len(kvs), kvs)
	}
	if kvs[0].Key != "endpoint" || kvs[0].Value != "http://example.com/rustfs/events" {
		t.Errorf("unexpected first kv: %+v", kvs[0])
	}
	if kvs[1].Key != "enable" || kvs[1].Value != "off" {
		t.Errorf("unexpected second kv: %+v", kvs[1])
	}
}

func TestGetConfigEscapedValue(t *testing.T) {
	client := newTestConfigAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(`notify_webhook:primary endpoint="http://example.com/a\"b\\c"`))
	})
	kvs, err := client.GetConfig("notify_webhook:primary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kvs) != 1 || kvs[0].Value != `http://example.com/a"b\c` {
		t.Errorf("unexpected kvs: %+v", kvs)
	}
}

func TestGetConfigNotFound(t *testing.T) {
	client := newTestConfigAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("config target 'missing' not found for subsystem 'notify_webhook'"))
	})
	_, err := client.GetConfig("notify_webhook:missing")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got: %v", err)
	}
}

func TestSetConfig(t *testing.T) {
	var gotPath string
	var gotBody []byte
	client := newTestConfigAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	kvs := []rustfs.ConfigKV{{Key: "endpoint", Value: "http://example.com/rustfs/events"}}
	if err := client.SetConfig("notify_webhook:primary", kvs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/set-config-kv") {
		t.Errorf("wrong path: %s", gotPath)
	}
	expected := `notify_webhook:primary endpoint="http://example.com/rustfs/events"`
	if string(gotBody) != expected {
		t.Errorf("unexpected body: %q, want %q", string(gotBody), expected)
	}
}

func TestSetConfigEscapesValue(t *testing.T) {
	var gotBody []byte
	client := newTestConfigAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	kvs := []rustfs.ConfigKV{{Key: "auth_token", Value: `a"b\c`}}
	if err := client.SetConfig("notify_webhook", kvs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `notify_webhook auth_token="a\"b\\c"`
	if string(gotBody) != expected {
		t.Errorf("unexpected body: %q, want %q", string(gotBody), expected)
	}
}

func TestDeleteConfig(t *testing.T) {
	var gotPath string
	var gotBody []byte
	client := newTestConfigAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	if err := client.DeleteConfig("notify_webhook:primary"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/del-config-kv") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if string(gotBody) != "notify_webhook:primary" {
		t.Errorf("unexpected body: %q", string(gotBody))
	}
}

func TestHelpConfig(t *testing.T) {
	client := newTestConfigAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rustfs/admin/v3/help-config-kv" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"subSys":          "notify_webhook",
			"description":     "publish bucket notifications to webhook endpoints",
			"multipleTargets": true,
			"keysHelp": []map[string]interface{}{
				{"key": "endpoint", "type": "url", "description": "webhook server endpoint", "optional": false, "multipleTargets": false},
				{"key": "enable", "type": "on|off", "description": "enable target", "optional": false, "multipleTargets": false},
			},
		})
	})
	kvs, err := client.HelpConfig("notify_webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kvs) != 2 || kvs[0].Key != "endpoint" || kvs[0].Value != "url" {
		t.Errorf("unexpected kvs: %+v", kvs)
	}
	if kvs[1].Key != "enable" || kvs[1].Value != "on|off" {
		t.Errorf("unexpected kvs: %+v", kvs)
	}
}
