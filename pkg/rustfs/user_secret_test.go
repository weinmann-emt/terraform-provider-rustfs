package rustfs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestAdminServer(t *testing.T, handler http.HandlerFunc) *RustfsAdmin {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"
	return &client
}

func TestSetUserSecretKey(t *testing.T) {
	var gotPath, gotBody string
	client := newTestAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		gotBody = strings.TrimSpace(buf.String())
		w.WriteHeader(http.StatusOK)
	})

	if err := client.SetUserSecretKey("alice", "new-secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/set-user-secret-key") || !strings.Contains(gotPath, "accessKey=alice") {
		t.Errorf("wrong request: %s", gotPath)
	}
	if !strings.Contains(gotBody, `"secret_key":"new-secret"`) {
		t.Errorf("wrong body: %s", gotBody)
	}
	if strings.Contains(gotBody, "secretKey") {
		t.Errorf("body must use snake_case secret_key, got: %s", gotBody)
	}
}

func TestSetUserStatus(t *testing.T) {
	var gotPath string
	client := newTestAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})

	if err := client.SetUserStatus("alice", "disabled"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/set-user-status") ||
		!strings.Contains(gotPath, "accessKey=alice") ||
		!strings.Contains(gotPath, "status=disabled") {
		t.Errorf("wrong request: %s", gotPath)
	}
}

func TestAttachPolicyToUser(t *testing.T) {
	var gotPath string
	client := newTestAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})

	if err := client.AttachPolicyToUser("alice", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/set-user-or-group-policy") ||
		!strings.Contains(gotPath, "userOrGroup=alice") ||
		!strings.Contains(gotPath, "policyName=readwrite") ||
		!strings.Contains(gotPath, "isGroup=false") {
		t.Errorf("wrong request: %s", gotPath)
	}
}
