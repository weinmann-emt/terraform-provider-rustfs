package rustfs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("group") != "developers" {
			t.Errorf("expected group=developers, got %s", r.URL.Query().Get("group"))
		}
		info := GroupInfo{
			Name:    "developers",
			Status:  "enabled",
			Members: []string{"alice", "bob"},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(info)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	info, err := client.GetGroup("developers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "developers" {
		t.Errorf("expected developers, got %s", info.Name)
	}
	if info.Status != "enabled" {
		t.Errorf("expected enabled, got %s", info.Status)
	}
	if len(info.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(info.Members))
	}
}

func TestUpdateGroupMembers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req GroupAddRemove
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		if req.Group != "developers" {
			t.Errorf("expected group=developers, got %s", req.Group)
		}
		if len(req.Members) != 2 {
			t.Errorf("expected 2 members, got %d", len(req.Members))
		}
		if req.IsRemove {
			t.Error("expected is_remove=false")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.UpdateGroupMembers(GroupAddRemove{
		Group:    "developers",
		Members:  []string{"alice", "bob"},
		IsRemove: false,
		Status:   "enabled",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.DeleteGroup("developers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetGroupStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Query().Get("group") != "developers" {
			t.Errorf("expected group=developers, got %s", r.URL.Query().Get("group"))
		}
		if r.URL.Query().Get("status") != "disabled" {
			t.Errorf("expected status=disabled, got %s", r.URL.Query().Get("status"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.SetGroupStatus("developers", "disabled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
