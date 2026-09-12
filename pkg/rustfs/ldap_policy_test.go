package rustfs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAttachLdapPolicy(t *testing.T) {
	var gotPath string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		gotBody, _ = io.ReadAll(r.Body)
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.AttachLDAPPolicy(LDAPPolicyAttachment{
		UserOrGroup: "uid=alice,ou=people,dc=example,dc=com",
		PolicyName:  "readwrite",
		IsGroup:     false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/idp/ldap/policy/attach") {
		t.Errorf("wrong request path: %s", gotPath)
	}

	var body map[string]any
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if policies, ok := body["policies"].([]any); !ok || len(policies) != 1 || policies[0] != "readwrite" {
		t.Errorf("wrong policies in body: %s", gotBody)
	}
	if user, ok := body["user"].(string); !ok || user != "uid=alice,ou=people,dc=example,dc=com" {
		t.Errorf("wrong user in body: %s", gotBody)
	}
	if _, hasGroup := body["group"]; hasGroup {
		t.Errorf("user attachment must not set group: %s", gotBody)
	}
}

func TestAttachLdapPolicyGroup(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.AttachLDAPPolicy(LDAPPolicyAttachment{
		UserOrGroup: "cn=engineers,ou=groups,dc=example,dc=com",
		PolicyName:  "readwrite",
		IsGroup:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if policies, ok := body["policies"].([]any); !ok || len(policies) != 1 || policies[0] != "readwrite" {
		t.Errorf("wrong policies in body: %s", gotBody)
	}
	if group, ok := body["group"].(string); !ok || group != "cn=engineers,ou=groups,dc=example,dc=com" {
		t.Errorf("wrong group in body: %s", gotBody)
	}
	if _, hasUser := body["user"]; hasUser {
		t.Errorf("group attachment must not set user: %s", gotBody)
	}
}

func TestDetachLdapPolicy(t *testing.T) {
	var gotPath string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		gotBody, _ = io.ReadAll(r.Body)
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.DetachLDAPPolicy(LDAPPolicyAttachment{
		UserOrGroup: "uid=alice,ou=people,dc=example,dc=com",
		PolicyName:  "readwrite",
		IsGroup:     false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/idp/ldap/policy/detach") {
		t.Errorf("wrong request path: %s", gotPath)
	}

	var body map[string]any
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if policies, ok := body["policies"].([]any); !ok || len(policies) != 1 || policies[0] != "readwrite" {
		t.Errorf("wrong policies in body: %s", gotBody)
	}
	if user, ok := body["user"].(string); !ok || user != "uid=alice,ou=people,dc=example,dc=com" {
		t.Errorf("wrong user in body: %s", gotBody)
	}
}
