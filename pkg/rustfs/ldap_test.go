package rustfs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateLDAPServiceAccount(t *testing.T) {
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	account := ServiceAccount{
		AccessKey:  "ldap-sa",
		SecretKey:  "s3cret",
		Name:       "ldap token",
		TargetUser: "uid=alice,ou=people,dc=example,dc=com",
		Policy:     "readwrite",
	}
	if err := client.CreateLDAPServiceAccount(account); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/rustfs/admin/v3/idp/ldap/add-service-account") {
		t.Errorf("wrong request path: %s", gotPath)
	}
	for _, want := range []string{
		`"accessKey":"ldap-sa"`,
		`"targetUser":"uid=alice,ou=people,dc=example,dc=com"`,
		`"policy":"readwrite"`,
	} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("missing %s in body: %s", want, gotBody)
		}
	}
}

func TestCreateLDAPServiceAccountError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`target user not exist`))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	account := ServiceAccount{
		AccessKey:  "ldap-sa",
		SecretKey:  "s3cret",
		TargetUser: "uid=ghost,ou=people,dc=example,dc=com",
	}
	err := client.CreateLDAPServiceAccount(account)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "target user not exist") {
		t.Errorf("unexpected error message: %v", err)
	}
}
