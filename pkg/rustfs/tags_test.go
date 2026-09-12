package rustfs

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBucketTagsSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/test-bucket" {
			t.Errorf("expected path /test-bucket, got %s", r.URL.Path)
		}
		if !r.URL.Query().Has("tagging") {
			t.Errorf("expected ?tagging query, got %s", r.URL.RawQuery)
		}
		body, _ := io.ReadAll(r.Body)
		var tagging bucketTagging
		if err := xml.Unmarshal(body, &tagging); err != nil {
			t.Errorf("expected valid tagging XML, got %s: %v", string(body), err)
		}
		if len(tagging.TagSet) != 2 {
			t.Errorf("expected 2 tags, got %d", len(tagging.TagSet))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	tags := map[string]string{
		"environment": "production",
		"team":        "platform",
	}
	err := client.SetBucketTagging("test-bucket", tags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBucketTagsSetSortedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "<Key>environment</Key><Value>production</Value>") {
			t.Errorf("unexpected XML body: %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.SetBucketTagging("test-bucket", map[string]string{
		"team":        "platform",
		"environment": "production",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBucketTagsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/test-bucket" {
			t.Errorf("expected path /test-bucket, got %s", r.URL.Path)
		}
		if !r.URL.Query().Has("tagging") {
			t.Errorf("expected ?tagging query, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<Tagging><TagSet><Tag><Key>environment</Key><Value>production</Value></Tag><Tag><Key>team</Key><Value>platform</Value></Tag></TagSet></Tagging>`))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	tags, err := client.GetBucketTagging("test-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags["environment"] != "production" {
		t.Errorf("expected environment=production, got %q", tags["environment"])
	}
	if tags["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", tags["team"])
	}
}

func TestBucketTagsGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<Error><Code>NoSuchTagSet</Code><Message>The TagSet does not exist</Message></Error>`))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	_, err := client.GetBucketTagging("test-bucket")
	if err == nil {
		t.Fatal("expected error for NoSuchTagSet, got nil")
	}
	if !strings.Contains(err.Error(), "NoSuchTagSet") {
		t.Errorf("expected NoSuchTagSet in error, got %q", err.Error())
	}
}

func TestBucketTagsRemove(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/test-bucket" {
			t.Errorf("expected path /test-bucket, got %s", r.URL.Path)
		}
		if !r.URL.Query().Has("tagging") {
			t.Errorf("expected ?tagging query, got %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.RemoveBucketTagging("test-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
