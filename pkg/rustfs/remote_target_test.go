package rustfs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddRemoteTarget(t *testing.T) {
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.RawQuery, "bucket=src-bkt") {
			t.Errorf("expected bucket query param, got %s", r.URL.RawQuery)
		}
		if strings.Contains(r.URL.RawQuery, "update=") {
			t.Errorf("create request must not carry update param, got %s", r.URL.RawQuery)
		}
		gotPath = r.URL.Path
		bodyBytes, _ := io.ReadAll(r.Body)
		gotBody = string(bodyBytes)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`"arn:minio:replication::abc:dst-bkt"`))
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	target := RemoteTarget{
		Type:         "replication",
		Endpoint:     "https://peer.example.com",
		Secure:       true,
		TargetBucket: "dst-bkt",
		Credentials: &RemoteTargetCredentials{
			AccessKey: "peeruser",
			SecretKey: "peersecret",
		},
	}
	arn, err := c.AddRemoteTarget("src-bkt", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arn != "arn:minio:replication::abc:dst-bkt" {
		t.Errorf("unexpected arn: %s", arn)
	}
	if !strings.Contains(gotPath, "/set-remote-target") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if !strings.Contains(gotBody, `"type":"replication"`) ||
		!strings.Contains(gotBody, `"targetbucket":"dst-bkt"`) ||
		!strings.Contains(gotBody, `"accessKey":"peeruser"`) {
		t.Errorf("wrong body: %s", gotBody)
	}
}

func TestUpdateRemoteTarget(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`"arn:minio:replication::abc:dst-bkt"`))
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	target := RemoteTarget{
		Arn:          "arn:minio:replication::abc:dst-bkt",
		Type:         "replication",
		Endpoint:     "https://peer.example.com",
		Secure:       true,
		TargetBucket: "dst-bkt",
		Credentials: &RemoteTargetCredentials{
			AccessKey: "peeruser",
			SecretKey: "peersecret",
		},
	}
	if _, err := c.AddRemoteTarget("src-bkt", target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotQuery, "update=true") {
		t.Errorf("expected update=true, got %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "creds=true") {
		t.Errorf("expected creds=true, got %s", gotQuery)
	}
}

func TestListRemoteTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/list-remote-targets") {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "bucket=src-bkt") {
			t.Errorf("expected bucket query param, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"sourcebucket":"src-bkt","arn":"arn:minio:replication:::peer:backup","type":"replication","endpoint":"https://peer.example.com","secure":true,"targetbucket":"backup"}]`))
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	targets, err := c.ListRemoteTargets("src-bkt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 || targets[0].Arn != "arn:minio:replication:::peer:backup" {
		t.Errorf("unexpected targets: %+v", targets)
	}
}

func TestDeleteRemoteTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/remove-remote-target") {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "bucket=src-bkt") ||
			!strings.Contains(r.URL.RawQuery, "arn=arn%3Aminio%3Areplication%3A%3Aabc%3Adst-bkt") {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})
	if err := c.DeleteRemoteTarget("src-bkt", "arn:minio:replication::abc:dst-bkt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
