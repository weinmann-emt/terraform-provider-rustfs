package rustfs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestS3Server(t *testing.T, handler http.HandlerFunc) *RustfsAdmin {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(srv.URL, "http://"),
	})
	return &c
}

func assertPublicAccessBlockQuery(t *testing.T, r *http.Request) {
	t.Helper()
	if _, ok := r.URL.Query()["publicAccessBlock"]; !ok {
		t.Errorf("expected publicAccessBlock subresource query, got raw query %q", r.URL.RawQuery)
	}
}

func TestSetBucketPublicAccessBlock(t *testing.T) {
	var gotPath, gotMethod, gotBody string
	client := newTestS3Server(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		assertPublicAccessBlockQuery(t, r)
		bodyBytes, _ := io.ReadAll(r.Body)
		gotBody = string(bodyBytes)
		w.WriteHeader(http.StatusOK)
	})

	err := client.SetBucketPublicAccessBlock("mybucket", &PublicAccessBlockConfiguration{
		BlockPublicAcls:   true,
		BlockPublicPolicy: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" || gotPath != "/mybucket" {
		t.Errorf("wrong request: %s %s", gotMethod, gotPath)
	}
	if !strings.Contains(gotBody, "<BlockPublicAcls>true</BlockPublicAcls>") ||
		!strings.Contains(gotBody, "<BlockPublicPolicy>true</BlockPublicPolicy>") {
		t.Errorf("body missing expected fields: %s", gotBody)
	}
}

func TestGetBucketPublicAccessBlock(t *testing.T) {
	client := newTestS3Server(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("wrong method: %s", r.Method)
		}
		assertPublicAccessBlockQuery(t, r)
		_, _ = w.Write([]byte(`<PublicAccessBlockConfiguration>` +
			`<BlockPublicAcls>true</BlockPublicAcls>` +
			`<IgnorePublicAcls>false</IgnorePublicAcls>` +
			`<BlockPublicPolicy>true</BlockPublicPolicy>` +
			`<RestrictPublicBuckets>false</RestrictPublicBuckets>` +
			`</PublicAccessBlockConfiguration>`))
	})

	cfg, err := client.GetBucketPublicAccessBlock("mybucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.BlockPublicAcls || !cfg.BlockPublicPolicy {
		t.Error("expected BlockPublicAcls and BlockPublicPolicy to be true")
	}
	if cfg.IgnorePublicAcls || cfg.RestrictPublicBuckets {
		t.Error("expected IgnorePublicAcls and RestrictPublicBuckets to be false")
	}
}

func TestGetBucketPublicAccessBlockNamespaced(t *testing.T) {
	client := newTestS3Server(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<PublicAccessBlockConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` +
			`<BlockPublicAcls>true</BlockPublicAcls>` +
			`<IgnorePublicAcls>true</IgnorePublicAcls>` +
			`<BlockPublicPolicy>false</BlockPublicPolicy>` +
			`<RestrictPublicBuckets>false</RestrictPublicBuckets>` +
			`</PublicAccessBlockConfiguration>`))
	})

	cfg, err := client.GetBucketPublicAccessBlock("mybucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.BlockPublicAcls || !cfg.IgnorePublicAcls {
		t.Error("expected BlockPublicAcls and IgnorePublicAcls to be true")
	}
	if cfg.BlockPublicPolicy || cfg.RestrictPublicBuckets {
		t.Error("expected BlockPublicPolicy and RestrictPublicBuckets to be false")
	}
}

func TestGetBucketPublicAccessBlockNotFound(t *testing.T) {
	client := newTestS3Server(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<Error><Code>NoSuchPublicAccessBlockConfiguration</Code></Error>`))
	})

	_, err := client.GetBucketPublicAccessBlock("mybucket")
	if err == nil {
		t.Fatal("expected an error for a missing public access block configuration")
	}
	if !strings.Contains(err.Error(), "NoSuchPublicAccessBlockConfiguration") {
		t.Errorf("expected NoSuchPublicAccessBlockConfiguration in error, got: %v", err)
	}
}

func TestDeleteBucketPublicAccessBlock(t *testing.T) {
	var gotPath, gotMethod string
	client := newTestS3Server(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		assertPublicAccessBlockQuery(t, r)
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.DeleteBucketPublicAccessBlock("mybucket"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "DELETE" || gotPath != "/mybucket" {
		t.Errorf("wrong request: %s %s", gotMethod, gotPath)
	}
}
