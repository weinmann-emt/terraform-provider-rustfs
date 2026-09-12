package rustfs

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestReplicationMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rustfs/admin/v3/replicationmetrics") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("bucket"); got != "mybucket" {
			t.Errorf("expected bucket=mybucket, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"Stats": {
				"arn:minio:replication::peer:backup": {
					"replicationCount": 9,
					"completedReplicationSize": 4096,
					"limitInBits": 1024,
					"currentBandwidth": 512.5,
					"failed": {
						"lastMinute": {"count": 1.0, "bytes": 128},
						"lastHour": {"count": 2.0, "bytes": 256},
						"totals": {"count": 3.0, "bytes": 900}
					},
					"failedReplicationSize": 900,
					"failedReplicationCount": 3
				}
			},
			"completedReplicationSize": 4096,
			"replicaSize": 128,
			"replicaCount": 2,
			"replicationCount": 9,
			"failed": {
				"lastMinute": {"count": 1.0, "bytes": 128},
				"lastHour": {"count": 2.0, "bytes": 256},
				"totals": {"count": 3.0, "bytes": 900}
			},
			"queued": {
				"curr": {"count": 4.0, "bytes": 1200.0},
				"avg": {"count": 2.0, "bytes": 600.0},
				"max": {"count": 5.0, "bytes": 1500.0},
				"peak": {"count": 5.0, "bytes": 1500.0}
			},
			"provider_available": true,
			"cluster_complete": true,
			"observed_node_count": 1,
			"expected_node_count": 1
		}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})

	metrics, err := c.ReplicationMetrics("mybucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.ReplicationCount != 9 {
		t.Errorf("unexpected replicationCount: %d", metrics.ReplicationCount)
	}
	if metrics.CompletedReplicationSize != 4096 {
		t.Errorf("unexpected completedReplicationSize: %d", metrics.CompletedReplicationSize)
	}
	if metrics.ReplicaCount != 2 {
		t.Errorf("unexpected replicaCount: %d", metrics.ReplicaCount)
	}
	if metrics.ReplicaSize != 128 {
		t.Errorf("unexpected replicaSize: %d", metrics.ReplicaSize)
	}
	if metrics.Failed.Totals.Count != 3.0 || metrics.Failed.Totals.Bytes != 900 {
		t.Errorf("unexpected failed totals: %+v", metrics.Failed.Totals)
	}
	if metrics.Queued.Curr.Count != 4.0 || metrics.Queued.Curr.Bytes != 1200.0 {
		t.Errorf("unexpected queued curr: %+v", metrics.Queued.Curr)
	}
	if !metrics.ProviderAvailable || !metrics.ClusterComplete {
		t.Errorf("unexpected source health: available=%v complete=%v", metrics.ProviderAvailable, metrics.ClusterComplete)
	}
	if metrics.ObservedNodeCount != 1 || metrics.ExpectedNodeCount != 1 {
		t.Errorf("unexpected node counts: observed=%d expected=%d", metrics.ObservedNodeCount, metrics.ExpectedNodeCount)
	}
	if len(metrics.Raw) == 0 {
		t.Error("expected raw JSON body to be captured")
	}

	target, ok := metrics.Stats["arn:minio:replication::peer:backup"]
	if !ok {
		t.Fatalf("expected stats for arn:minio:replication::peer:backup, got %v", metrics.Stats)
	}
	if target.ReplicationCount != 9 {
		t.Errorf("unexpected target replicationCount: %d", target.ReplicationCount)
	}
	if target.CompletedReplicationSize != 4096 {
		t.Errorf("unexpected target completedReplicationSize: %d", target.CompletedReplicationSize)
	}
	if target.LimitInBits != 1024 {
		t.Errorf("unexpected target limitInBits: %d", target.LimitInBits)
	}
	if target.CurrentBandwidth != 512.5 {
		t.Errorf("unexpected target currentBandwidth: %f", target.CurrentBandwidth)
	}
	if target.Failed.Totals.Count != 3.0 || target.FailedReplicationSize != 900 || target.FailedReplicationCount != 3 {
		t.Errorf("unexpected target failed stats: %+v", target)
	}
}

func TestReplicationMetricsNoReplicationConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>ReplicationConfigurationNotFoundError</Code><Message>replication not found</Message></Error>`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})

	metrics, err := c.ReplicationMetrics("mybucket")
	if err != nil {
		t.Fatalf("expected no error when replication is not configured, got: %v", err)
	}
	if len(metrics.Stats) != 0 {
		t.Errorf("expected empty stats, got %v", metrics.Stats)
	}
}

func TestReplicationMetricsMissingBucket(t *testing.T) {
	c := New(&RustfsAdminConfig{AccessKey: "admin", AccessSecret: "secret"})
	if _, err := c.ReplicationMetrics(""); err == nil {
		t.Fatal("expected error for empty bucket, got nil")
	}
}

func TestReplicationMetricsBucketNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchBucket</Code><Message>The specified bucket does not exist</Message></Error>`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})

	if _, err := c.ReplicationMetrics("nope"); err == nil {
		t.Fatal("expected error for nonexistent bucket, got nil")
	}
}

func TestReplicationMetricsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("boom")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})

	if _, err := c.ReplicationMetrics("mybucket"); err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestReplicationMetricsSendsBucketQuery(t *testing.T) {
	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"Stats":{},"completedReplicationSize":0,"replicaSize":0,"replicaCount":0,"replicationCount":0,"failed":{"lastMinute":{"count":0,"bytes":0},"lastHour":{"count":0,"bytes":0},"totals":{"count":0,"bytes":0}},"queued":{"curr":{"count":0,"bytes":0},"avg":{"count":0,"bytes":0},"max":{"count":0,"bytes":0},"peak":{"count":0,"bytes":0}},"provider_available":true,"cluster_complete":true,"observed_node_count":1,"expected_node_count":1}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := New(&RustfsAdminConfig{
		AccessKey:    "admin",
		AccessSecret: "secret",
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
	})

	if _, err := c.ReplicationMetrics("quoted bucket!?&"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Get("bucket") != "quoted bucket!?&" {
		t.Errorf("bucket query param not sent correctly: %v", got)
	}
}
