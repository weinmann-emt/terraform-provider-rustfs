package rustfs_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestUserPolicyAttachmentAttach(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	client := newTestAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	if err := client.AttachUserPolicy("alice", "readonly"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/idp/builtin/policy/attach") {
		t.Errorf("wrong request path: %s", gotPath)
	}
	body := string(gotBody)
	if !strings.Contains(body, `"user":"alice"`) || !strings.Contains(body, `"readonly"`) {
		t.Errorf("wrong request body: %s", body)
	}
}

func TestUserPolicyAttachmentDetach(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	client := newTestAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	if err := client.DetachUserPolicy("alice", "readonly"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/idp/builtin/policy/detach") {
		t.Errorf("wrong request path: %s", gotPath)
	}
	body := string(gotBody)
	if !strings.Contains(body, `"user":"alice"`) || !strings.Contains(body, `"readonly"`) {
		t.Errorf("wrong request body: %s", body)
	}
}

func TestUserPolicyAttachmentError(t *testing.T) {
	client := newTestAdminServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	if err := client.AttachUserPolicy("alice", "readonly"); err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}
