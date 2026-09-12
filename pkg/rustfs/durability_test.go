package rustfs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetBucketDurability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rustfs/admin/v3/bucket-durability/mybucket") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"bucket":"mybucket","mode":"relaxed"}`))
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	d, err := c.GetBucketDurability("mybucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Bucket != "mybucket" || d.Mode != "relaxed" {
		t.Errorf("unexpected durability: %+v", d)
	}
}

func TestGetBucketDurabilityInheritsNullMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"bucket":"mybucket","mode":null}`))
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	d, err := c.GetBucketDurability("mybucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Mode != "" {
		t.Errorf("expected empty mode for inherited durability, got %q", d.Mode)
	}
}

func TestSetBucketDurability(t *testing.T) {
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"bucket":"mybucket","mode":"strict"}`))
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	d, err := c.SetBucketDurability("mybucket", BucketDurability{Bucket: "mybucket", Mode: "strict"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/bucket-durability/mybucket") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if !strings.Contains(gotBody, `"mode":"strict"`) {
		t.Errorf("wrong body: %s", gotBody)
	}
	if d.Bucket != "mybucket" || d.Mode != "strict" {
		t.Errorf("unexpected readback: %+v", d)
	}
}

func TestDeleteBucketDurability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/bucket-durability/mybucket") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	if err := c.DeleteBucketDurability("mybucket"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
