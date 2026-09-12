package rustfs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSiteReplicationAdd(t *testing.T) {
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = readBody(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c := testSiteReplicationClient(server)

	site := SiteReplicationSite{
		Name:          "peer",
		Endpoint:      "http://localhost:9002",
		AccessKey:     "peeradmin",
		SecretKey:     "peersecret",
		SkipTLSVerify: false,
	}
	if err := c.SiteReplicationAdd(site); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/site-replication/add") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if !strings.Contains(gotBody, `"endpoints":"http://localhost:9002"`) {
		t.Errorf("expected endpoints field, got: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"accessKey":"peeradmin"`) {
		t.Errorf("expected accessKey field, got: %s", gotBody)
	}
	var payload []SiteReplicationSite
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("add body must be a JSON array: %v", err)
	}
	if len(payload) != 1 || payload[0].Name != "peer" {
		t.Errorf("unexpected payload: %+v", payload)
	}
}

func TestSiteReplicationEdit(t *testing.T) {
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = readBody(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c := testSiteReplicationClient(server)

	site := SiteReplicationSite{
		Name:          "peer",
		Endpoint:      "http://localhost:9003",
		SkipTLSVerify: true,
	}
	if err := c.SiteReplicationEdit(site); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/site-replication/edit") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if !strings.Contains(gotBody, `"endpoints":"http://localhost:9003"`) {
		t.Errorf("expected endpoints field, got: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"skipTlsVerify":true`) {
		t.Errorf("expected skipTlsVerify field, got: %s", gotBody)
	}
}

func TestSiteReplicationRemove(t *testing.T) {
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = readBody(r)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c := testSiteReplicationClient(server)

	if err := c.SiteReplicationRemove([]string{"peer"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/site-replication/remove") {
		t.Errorf("wrong path: %s", gotPath)
	}
	var body struct {
		Sites []string `json:"sites"`
	}
	if err := json.Unmarshal([]byte(gotBody), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if len(body.Sites) != 1 || body.Sites[0] != "peer" {
		t.Errorf("wrong body: %s", gotBody)
	}
}

func TestSiteReplicationInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/site-replication/info") {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"enabled":true,"name":"local","sites":[{"name":"peer","endpoint":"http://localhost:9002","deploymentID":"dep1"}]}`))
	}))
	defer server.Close()
	c := testSiteReplicationClient(server)

	info, err := c.SiteReplicationInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Enabled || len(info.Sites) != 1 || info.Sites[0].Name != "peer" || info.Sites[0].Endpoint != "http://localhost:9002" {
		t.Errorf("unexpected info: %+v", info)
	}
}

func TestSiteReplicationResyncOp(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"op":"start","id":"abc","status":"success","state":"pending"}`))
	}))
	defer server.Close()
	c := testSiteReplicationClient(server)

	resync, err := c.SiteReplicationResyncOp("start", SiteReplicationPeer{Name: "peer", DeploymentID: "dep1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotPath, "/site-replication/resync/op") {
		t.Errorf("wrong path: %s", gotPath)
	}
	if !strings.Contains(gotPath, "operation=start") {
		t.Errorf("expected operation=start, got: %s", gotPath)
	}
	if resync.OpType != "start" || resync.Status != "success" {
		t.Errorf("unexpected resync: %+v", resync)
	}
}

func testSiteReplicationClient(server *httptest.Server) RustfsAdmin {
	c := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	c.accessSecret = "secret"
	return c
}

func readBody(r *http.Request) string {
	b, _ := io.ReadAll(r.Body)
	return string(b)
}
