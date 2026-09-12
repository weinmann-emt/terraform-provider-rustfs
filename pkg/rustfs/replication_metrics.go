package rustfs

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/url"
)

// ReplicationRStat mirrors minio-go replication.RStat: a single (count, bytes)
// measurement window.
type ReplicationRStat struct {
	Count float64 `json:"count"`
	Bytes int64   `json:"bytes"`
}

// ReplicationTimedErrStats mirrors minio-go replication.TimedErrStats: rolling
// failure windows plus cumulative totals.
type ReplicationTimedErrStats struct {
	LastMinute ReplicationRStat `json:"lastMinute"`
	LastHour   ReplicationRStat `json:"lastHour"`
	Totals     ReplicationRStat `json:"totals"`
}

// ReplicationQStat mirrors minio-go replication.QStat.
type ReplicationQStat struct {
	Count float64 `json:"count"`
	Bytes float64 `json:"bytes"`
}

// ReplicationInQueueMetric mirrors minio-go replication.InQueueMetric. Both
// `max` (MinIO server tag) and `peak` (minio-go tag) carry the queue peak.
type ReplicationInQueueMetric struct {
	Curr ReplicationQStat `json:"curr"`
	Avg  ReplicationQStat `json:"avg"`
	Max  ReplicationQStat `json:"max"`
	Peak ReplicationQStat `json:"peak"`
}

// ReplicationTargetMetric mirrors minio-go replication.TargetMetrics for a
// single replication target (ARN).
type ReplicationTargetMetric struct {
	ReplicationCount         int64                    `json:"replicationCount"`
	CompletedReplicationSize int64                    `json:"completedReplicationSize"`
	LimitInBits              int64                    `json:"limitInBits"`
	CurrentBandwidth         float64                  `json:"currentBandwidth"`
	Failed                   ReplicationTimedErrStats `json:"failed"`
	FailedReplicationSize    int64                    `json:"failedReplicationSize"`
	FailedReplicationCount   int64                    `json:"failedReplicationCount"`
}

// ReplicationMetrics is the minio-go replication.Metrics wire shape returned
// by GET /rustfs/admin/v3/replicationmetrics. The trailing snake_case fields
// are RustFS source-health extension keys.
type ReplicationMetrics struct {
	Stats                    map[string]ReplicationTargetMetric `json:"Stats"`
	CompletedReplicationSize int64                              `json:"completedReplicationSize"`
	ReplicaSize              int64                              `json:"replicaSize"`
	ReplicaCount             int64                              `json:"replicaCount"`
	ReplicationCount         int64                              `json:"replicationCount"`
	Failed                   ReplicationTimedErrStats           `json:"failed"`
	Queued                   ReplicationInQueueMetric           `json:"queued"`
	ProviderAvailable        bool                               `json:"provider_available"`
	ClusterComplete          bool                               `json:"cluster_complete"`
	ObservedNodeCount        uint32                             `json:"observed_node_count"`
	ExpectedNodeCount        uint32                             `json:"expected_node_count"`

	// Raw holds the unparsed response body for consumers that want the
	// exact JSON as returned by the server.
	Raw json.RawMessage `json:"-"`
}

type s3ErrorXML struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

// ReplicationMetrics returns replication transfer statistics for a single
// bucket. When the bucket has no replication configuration configured the
// server reports ReplicationConfigurationNotFoundError, which is treated as
// an empty result rather than an error.
func (c *RustfsAdmin) ReplicationMetrics(bucket string) (ReplicationMetrics, error) {
	if bucket == "" {
		return ReplicationMetrics{}, errors.New("bucket is required")
	}
	reqData := RequestData{
		Method:  "GET",
		RelPath: "replicationmetrics",
		QueryValues: url.Values{
			"bucket": {bucket},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		if replicationErrorCode(err) == "ReplicationConfigurationNotFoundError" {
			return ReplicationMetrics{}, nil
		}
		return ReplicationMetrics{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReplicationMetrics{}, err
	}
	var metrics ReplicationMetrics
	metrics.Raw = json.RawMessage(body)
	if err := json.Unmarshal(body, &metrics); err != nil {
		return ReplicationMetrics{}, err
	}
	return metrics, nil
}

func replicationErrorCode(err error) string {
	var e s3ErrorXML
	if xmlErr := xml.Unmarshal([]byte(err.Error()), &e); xmlErr != nil {
		return ""
	}
	return e.Code
}
