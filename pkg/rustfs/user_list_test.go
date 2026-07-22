package rustfs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		users := []UserInfo{
			{AccessKey: "alice", Status: "enabled", Policy: "readwrite"},
			{AccessKey: "bob", Status: "disabled", Policy: "readonly"},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	users, err := client.ListUsers("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].AccessKey != "alice" {
		t.Errorf("expected alice, got %s", users[0].AccessKey)
	}
	if users[1].Status != "disabled" {
		t.Errorf("expected disabled, got %s", users[1].Status)
	}
}

func TestListUsersWithBucket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucket := r.URL.Query().Get("bucket")
		if bucket != "my-bucket" {
			t.Errorf("expected bucket=my-bucket, got %s", bucket)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]UserInfo{})
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	users, err := client.ListUsers("my-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected 0 users, got %d", len(users))
	}
}
