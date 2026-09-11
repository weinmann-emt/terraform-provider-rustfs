package rustfs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newAuditTargetTestServer(t *testing.T, handler http.HandlerFunc) *RustfsAdmin {
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

func TestListAuditTargets(t *testing.T) {
	client := newAuditTargetTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/audit/target/list" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"audit_endpoints":[{"account_id":"audit","service":"webhook","status":"offline","health_state":"offline","health_reason":"not_loaded_in_runtime","source":"config"}]}`))
	})
	targets, err := client.ListAuditTargets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	target := targets[0]
	if target.AccountID != "audit" || target.Service != "webhook" {
		t.Errorf("wrong target: %+v", target)
	}
	if target.Source != "config" || target.Status != "offline" {
		t.Errorf("wrong target metadata: %+v", target)
	}
}

func TestSetAuditTarget(t *testing.T) {
	client := newAuditTargetTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/audit/target/audit_webhook/audit" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var payload auditTargetSetBody
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("invalid body: %v", err)
		}
		if len(payload.KeyValues) != 3 {
			t.Fatalf("expected 3 key values, got %d", len(payload.KeyValues))
		}
		kv := map[string]string{}
		for _, item := range payload.KeyValues {
			kv[item.Key] = item.Value
		}
		if kv["endpoint"] != "https://hooks.example/webhook" || kv["auth_token"] != "s3cr3t" || kv["comment"] != "audit logs" {
			t.Errorf("wrong body: %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
	})
	err := client.SetAuditTarget("audit_webhook", "audit", []AuditTargetKeyValue{
		{Key: "endpoint", Value: "https://hooks.example/webhook"},
		{Key: "auth_token", Value: "s3cr3t"},
		{Key: "comment", Value: "audit logs"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResetAuditTarget(t *testing.T) {
	client := newAuditTargetTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/audit/target/audit_webhook/audit/reset" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	if err := client.ResetAuditTarget("audit_webhook", "audit"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAuditTargetError(t *testing.T) {
	client := newAuditTargetTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"Code":"InvalidArgument","Message":"endpoint is required"}`))
	})
	err := client.SetAuditTarget("audit_webhook", "audit", []AuditTargetKeyValue{{Key: "comment", Value: "x"}})
	if err == nil {
		t.Fatal("expected an error for a rejected request")
	}
}
