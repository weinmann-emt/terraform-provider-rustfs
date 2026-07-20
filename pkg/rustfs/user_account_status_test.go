package rustfs

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateUserAccountWithStatus(t *testing.T) {
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path != "/rustfs/admin/v3/user-info" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		status := r.URL.Query().Get("status")
		accessKey := r.URL.Query().Get("accessKey")
		if status != "disabled" {
			t.Errorf("expected status=disabled, got %s", status)
		}
		if accessKey != "testuser" {
			t.Errorf("expected accessKey=testuser, got %s", accessKey)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	// Pass no policy so only user-info is called
	err := client.UpdateUserAccount(UserAccount{
		AccessKey: "testuser",
		SecretKey: "newsecret",
		Status:    "disabled",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}
}

func TestUpdateUserAccountWithoutStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		if status != "" {
			t.Errorf("expected empty status, got %s", status)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.UpdateUserAccount(UserAccount{
		AccessKey: "testuser",
		SecretKey: "newsecret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
