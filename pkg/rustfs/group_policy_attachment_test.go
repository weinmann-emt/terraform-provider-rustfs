package rustfs

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAttachGroupPolicyAttachment(t *testing.T) {
	var gotMethod string
	var gotQuery url.Values
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	if err := client.AttachGroupPolicy("developers", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/set-user-or-group-policy") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if gotQuery.Get("userOrGroup") != "developers" {
		t.Errorf("expected userOrGroup=developers, got %s", gotQuery.Get("userOrGroup"))
	}
	if gotQuery.Get("policyName") != "readwrite" {
		t.Errorf("expected policyName=readwrite, got %s", gotQuery.Get("policyName"))
	}
	if gotQuery.Get("isGroup") != "true" {
		t.Errorf("expected isGroup=true, got %s", gotQuery.Get("isGroup"))
	}
}

func TestDetachGroupPolicyAttachment(t *testing.T) {
	var gotMethod string
	var gotQuery url.Values
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	if err := client.DetachGroupPolicy("developers", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/set-user-or-group-policy") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if gotQuery.Get("userOrGroup") != "developers" {
		t.Errorf("expected userOrGroup=developers, got %s", gotQuery.Get("userOrGroup"))
	}
	if _, ok := gotQuery["policyName"]; !ok {
		t.Error("expected policyName key in query")
	}
	if gotQuery.Get("policyName") != "" {
		t.Errorf("expected empty policyName for detach, got %s", gotQuery.Get("policyName"))
	}
	if gotQuery.Get("isGroup") != "true" {
		t.Errorf("expected isGroup=true, got %s", gotQuery.Get("isGroup"))
	}
}
