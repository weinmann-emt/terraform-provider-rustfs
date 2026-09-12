package rustfs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newModuleSwitchTestServer(t *testing.T, handler http.HandlerFunc) *RustfsAdmin {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(srv.URL, "http://"),
	})
	return &c
}

func TestGetModuleSwitches(t *testing.T) {
	client := newModuleSwitchTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/module-switches" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"notify_enabled":true,"audit_enabled":false,"persisted_notify_enabled":true,"persisted_audit_enabled":false,"notify_source":"console","audit_source":"console","admin_discovery":{"runtimeCapabilities":"/x","clusterSnapshot":"/y","extensionsCatalog":"/z"}}`))
	})
	state, err := client.GetModuleSwitches()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.NotifyEnabled {
		t.Errorf("expected notify_enabled true, got %v", state.NotifyEnabled)
	}
	if state.AuditEnabled {
		t.Errorf("expected audit_enabled false, got %v", state.AuditEnabled)
	}
	if state.NotifySource != ModuleSwitchSourceConsole {
		t.Errorf("expected notify_source console, got %q", state.NotifySource)
	}
	if state.AdminDiscovery == nil || state.AdminDiscovery.RuntimeCapabilities != "/x" {
		t.Errorf("expected admin_discovery.runtimeCapabilities /x, got %+v", state.AdminDiscovery)
	}
}

func TestSetModuleSwitches(t *testing.T) {
	var called bool
	client := newModuleSwitchTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/module-switches" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		var body ModuleSwitchUpdate
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("invalid body: %v", err)
		}
		if !body.NotifyEnabled || body.AuditEnabled {
			t.Errorf("wrong body: %+v", body)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"notify_enabled":true,"audit_enabled":false,"persisted_notify_enabled":true,"persisted_audit_enabled":false,"notify_source":"console","audit_source":"console"}`))
	})
	state, err := client.SetModuleSwitches(ModuleSwitchUpdate{NotifyEnabled: true, AuditEnabled: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler was not invoked")
	}
	if !state.NotifyEnabled {
		t.Errorf("expected returned notify_enabled true, got %v", state.NotifyEnabled)
	}
}
