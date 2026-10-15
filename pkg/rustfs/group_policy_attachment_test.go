package rustfs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

// fakeGroupPolicyServer mimics the admin API: GET group returns the mapping as
// a comma-separated string, set-user-or-group-policy replaces it wholesale.
type fakeGroupPolicyServer struct {
	mu       sync.Mutex
	policies map[string]string
	sets     []string
}

func (f *fakeGroupPolicyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	q := r.URL.Query()
	switch {
	case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/group"):
		_ = json.NewEncoder(w).Encode(GroupInfo{Name: q.Get("group"), Policy: f.policies[q.Get("group")]})
	case r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/set-user-or-group-policy"):
		if q.Get("isGroup") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f.policies[q.Get("userOrGroup")] = q.Get("policyName")
		f.sets = append(f.sets, q.Get("policyName"))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func newFakeGroupPolicyClient(t *testing.T, initial map[string]string) (*RustfsAdmin, *fakeGroupPolicyServer) {
	t.Helper()
	fake := &fakeGroupPolicyServer{policies: initial}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	client := New(&RustfsAdminConfig{
		Endpoint:     server.Listener.Addr().String(),
		AccessKey:    "admin",
		AccessSecret: "secret",
	})
	return &client, fake
}

func TestAttachGroupPolicyKeepsExisting(t *testing.T) {
	client, fake := newFakeGroupPolicyClient(t, map[string]string{"developers": "readonly"})

	if err := client.AttachGroupPolicy("developers", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := fake.policies["developers"]; got != "readonly,readwrite" {
		t.Errorf("expected readonly,readwrite, got %q", got)
	}
}

func TestAttachGroupPolicyAlreadyAttached(t *testing.T) {
	client, fake := newFakeGroupPolicyClient(t, map[string]string{"developers": "readonly, readwrite"})

	if err := client.AttachGroupPolicy("developers", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.sets) != 0 {
		t.Errorf("expected no write, got %v", fake.sets)
	}
}

func TestDetachGroupPolicyKeepsOthers(t *testing.T) {
	client, fake := newFakeGroupPolicyClient(t, map[string]string{"developers": "readonly,readwrite,diag"})

	if err := client.DetachGroupPolicy("developers", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := fake.policies["developers"]; got != "readonly,diag" {
		t.Errorf("expected readonly,diag, got %q", got)
	}
}

func TestDetachGroupPolicyLast(t *testing.T) {
	client, fake := newFakeGroupPolicyClient(t, map[string]string{"developers": "readwrite"})

	if err := client.DetachGroupPolicy("developers", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(fake.sets, []string{""}) {
		t.Errorf("expected a single clearing write, got %v", fake.sets)
	}
}

func TestDetachGroupPolicyNotAttached(t *testing.T) {
	client, fake := newFakeGroupPolicyClient(t, map[string]string{"developers": "readonly"})

	if err := client.DetachGroupPolicy("developers", "readwrite"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.sets) != 0 {
		t.Errorf("expected no write, got %v", fake.sets)
	}
}

func TestAttachGroupPolicyConcurrent(t *testing.T) {
	client, fake := newFakeGroupPolicyClient(t, map[string]string{})

	const n = 10
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := client.AttachGroupPolicy("developers", fmt.Sprintf("p%d", i)); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	got := splitPolicies(fake.policies["developers"])
	if len(got) != n {
		t.Errorf("expected %d policies, got %v", n, got)
	}
}

func TestGetGroupPolicies(t *testing.T) {
	client, _ := newFakeGroupPolicyClient(t, map[string]string{"developers": "a, b,,c"})

	got, err := client.GetGroupPolicies("developers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("unexpected policies: %v", got)
	}
}
