package rustfs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newKmsTestClient(t *testing.T, handler http.HandlerFunc) *RustfsAdmin {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Amz-Date") == "" {
			t.Error("expected signed request headers")
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"
	return &client
}

func readBody(t *testing.T, body io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	return string(b)
}

func TestCreateKmsKey(t *testing.T) {
	client := newKmsTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/keys" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		var body struct {
			Tags map[string]string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("invalid body: %v", err)
		}
		if body.Tags["name"] != "mykey" {
			t.Errorf("wrong name tag: %s", body.Tags["name"])
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"key created successfully","key_id":"mykey","key_metadata":{"key_id":"mykey","key_state":"Enabled","key_usage":"EncryptDecrypt","creation_date":"2026-09-07T00:00:00Z","origin":"KMS","key_manager":"CUSTOMER","tags":{"name":"mykey"}}}`))
	})

	key, err := client.CreateKmsKey("mykey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.KeyID != "mykey" {
		t.Errorf("wrong key id: %s", key.KeyID)
	}
	if key.KeyState != "Enabled" {
		t.Errorf("wrong key state: %s", key.KeyState)
	}
	if key.Tags["name"] != "mykey" {
		t.Errorf("missing name tag: %+v", key.Tags)
	}
}

func TestDescribeKmsKey(t *testing.T) {
	client := newKmsTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/keys/k1" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"Key described successfully","key_metadata":{"key_id":"k1","key_state":"Disabled","key_usage":"EncryptDecrypt","creation_date":"2026-09-07T00:00:00Z","origin":"KMS","key_manager":"CUSTOMER","tags":{"name":"k1"}},"impact":null}`))
	})

	key, err := client.DescribeKmsKey("k1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.KeyID != "k1" {
		t.Errorf("wrong key id: %s", key.KeyID)
	}
	if key.KeyState != "Disabled" {
		t.Errorf("wrong key state: %s", key.KeyState)
	}
}

func TestListKmsKeys(t *testing.T) {
	client := newKmsTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/keys" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"keys listed successfully","keys":[{"key_id":"k1","status":"Active","created_at":"2026-09-07T00:00:00Z","created_by":"local-kms"},{"key_id":"k2","status":"Active","created_at":"2026-09-07T00:00:00Z"}],"truncated":false,"next_marker":null}`))
	})

	keys, err := client.ListKmsKeys()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].KeyID != "k1" || keys[1].KeyID != "k2" {
		t.Errorf("wrong keys: %+v", keys)
	}
	if keys[0].CreatedBy == nil || *keys[0].CreatedBy != "local-kms" {
		t.Errorf("wrong created_by: %+v", keys[0].CreatedBy)
	}
}

func TestEnableKmsKey(t *testing.T) {
	client := newKmsTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/keys/enable" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		body := readBody(t, r.Body)
		var req struct {
			KeyID string `json:"key_id"`
		}
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			t.Errorf("invalid body: %v", err)
		}
		if req.KeyID != "mykey" {
			t.Errorf("wrong key id: %s", req.KeyID)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"key enabled successfully","key_id":"mykey","key_metadata":null}`))
	})

	if err := client.EnableKmsKey("mykey"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDisableKmsKey(t *testing.T) {
	client := newKmsTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/keys/disable" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		var req struct {
			KeyID string `json:"key_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("invalid body: %v", err)
		}
		if req.KeyID != "mykey" {
			t.Errorf("wrong key id: %s", req.KeyID)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"key disabled successfully","key_id":"mykey","key_metadata":null}`))
	})

	if err := client.DisableKmsKey("mykey"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRotateKmsKey(t *testing.T) {
	client := newKmsTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/keys/rotate" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		var req struct {
			KeyID string `json:"key_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("invalid body: %v", err)
		}
		if req.KeyID != "mykey" {
			t.Errorf("wrong key id: %s", req.KeyID)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"key rotated successfully","key_id":"mykey","key_metadata":{"key_id":"mykey","key_state":"Enabled","creation_date":"2026-09-07T01:00:00Z","origin":"KMS","key_manager":"CUSTOMER"}}`))
	})

	key, err := client.RotateKmsKey("mykey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.KeyID != "mykey" {
		t.Errorf("wrong key id: %s", key.KeyID)
	}
}

func TestDeleteKmsKey(t *testing.T) {
	client := newKmsTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/rustfs/admin/v3/kms/keys/delete" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		var req struct {
			KeyID string `json:"key_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("invalid body: %v", err)
		}
		if req.KeyID != "mykey" {
			t.Errorf("wrong key id: %s", req.KeyID)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"key deleted successfully","key_id":"mykey","deletion_date":"2026-10-07T00:00:00Z","impact":null}`))
	})

	if err := client.DeleteKmsKey("mykey"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
