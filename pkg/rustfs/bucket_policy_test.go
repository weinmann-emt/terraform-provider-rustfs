package rustfs_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

const testBucketPolicy = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::test-bucket/*"]}]}`

func newBucketPolicyTestClient(t *testing.T, handler http.HandlerFunc) *rustfs.RustfsAdmin {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	endpoint := strings.TrimPrefix(srv.URL, "http://")
	config := rustfs.RustfsAdminConfig{
		AccessKey:    "test",
		AccessSecret: "test",
		Endpoint:     endpoint,
		Ssl:          false,
	}
	client := rustfs.New(&config)
	return &client
}

func TestBucketPolicySetGetRemove(t *testing.T) {
	var mu sync.Mutex
	stored := ""

	server := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.URL.Path != "/test-bucket" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("policy"); got != "" {
			t.Errorf("expected policy query param, got %q", r.URL.Query().Encode())
		}

		switch r.Method {
		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read body: %v", err)
			}
			stored = strings.TrimSpace(string(body))
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			if stored == "" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, stored)
		case http.MethodDelete:
			stored = ""
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

	client := newBucketPolicyTestClient(t, server)

	if err := client.SetBucketPolicy("test-bucket", testBucketPolicy); err != nil {
		t.Fatalf("SetBucketPolicy: %v", err)
	}

	got, err := client.GetBucketPolicy("test-bucket")
	if err != nil {
		t.Fatalf("GetBucketPolicy: %v", err)
	}
	if got != testBucketPolicy {
		t.Fatalf("GetBucketPolicy = %q, want %q", got, testBucketPolicy)
	}

	if err := client.RemoveBucketPolicy("test-bucket"); err != nil {
		t.Fatalf("RemoveBucketPolicy: %v", err)
	}

	got, err = client.GetBucketPolicy("test-bucket")
	if err != nil {
		t.Fatalf("GetBucketPolicy after remove: %v", err)
	}
	if got != "" {
		t.Fatalf("GetBucketPolicy after remove = %q, want empty", got)
	}
}

func TestBucketPolicyGetWhenNoneExists(t *testing.T) {
	server := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	client := newBucketPolicyTestClient(t, server)

	got, err := client.GetBucketPolicy("missing-bucket")
	if err != nil {
		t.Fatalf("GetBucketPolicy: %v", err)
	}
	if got != "" {
		t.Fatalf("GetBucketPolicy = %q, want empty for missing policy", got)
	}
}

func TestBucketPolicySetError(t *testing.T) {
	server := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "access denied")
	}

	client := newBucketPolicyTestClient(t, server)

	if err := client.SetBucketPolicy("test-bucket", testBucketPolicy); err == nil {
		t.Fatal("expected SetBucketPolicy to return an error for 403")
	}
}
