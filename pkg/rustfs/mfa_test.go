package rustfs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestMfaServer(t *testing.T, handler http.HandlerFunc) RustfsAdmin {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := New(&RustfsAdminConfig{
		Endpoint:     server.Listener.Addr().String(),
		AccessKey:    "admin",
		AccessSecret: "secret",
	})
	return client
}

func TestReadUserMFA(t *testing.T) {
	var gotPath string
	client := newTestMfaServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_key":"alice","enabled":true,"activated_at":"2024-01-01T00:00:00Z","recovery_codes_remaining":5}`))
	})

	status, err := client.ReadUserMFA("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/user/mfa") || !strings.Contains(gotPath, "accessKey=alice") {
		t.Errorf("wrong request: %s", gotPath)
	}
	if !status.Enabled {
		t.Error("expected enabled=true")
	}
	if status.AccessKey != "alice" {
		t.Errorf("expected access_key alice, got %s", status.AccessKey)
	}
	if status.RecoveryCodesRemaining != 5 {
		t.Errorf("expected 5 recovery codes remaining, got %d", status.RecoveryCodesRemaining)
	}
}

func TestReadUserMFAWhenDisabled(t *testing.T) {
	client := newTestMfaServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_key":"alice","enabled":false,"recovery_codes_remaining":0}`))
	})

	status, err := client.ReadUserMFA("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Enabled {
		t.Error("expected enabled=false")
	}
	if status.ActivatedAt != "" {
		t.Errorf("expected empty activated_at, got %s", status.ActivatedAt)
	}
}

func TestClearUserMFA(t *testing.T) {
	var gotMethod, gotPath string
	client := newTestMfaServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})

	if err := client.ClearUserMFA("alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/user/mfa") || !strings.Contains(gotPath, "accessKey=alice") {
		t.Errorf("wrong request: %s", gotPath)
	}
}
