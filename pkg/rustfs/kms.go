package rustfs

import (
	"context"
	"encoding/json"
)

// KMSStatus is the response of GET /rustfs/admin/v3/kms/status.
//
// The optional fields are only present once a KMS backend has been configured
// and started; an unconfigured server returns the generic "KMS service not
// initialized" error instead of a document.
type KMSStatus struct {
	BackendType   string                  `json:"backend_type"`
	BackendStatus string                  `json:"backend_status"`
	CacheEnabled  bool                    `json:"cache_enabled"`
	CacheStats    *KMSCacheStats          `json:"cache_stats"`
	DefaultKeyID  *string                 `json:"default_key_id"`
	Capabilities  *KMSBackendCapabilities `json:"capabilities"`
	ClusterConfig *KMSClusterConfigStatus `json:"cluster_config"`
}

// KMSCacheStats is the per-server KMS key cache statistics block.
type KMSCacheStats struct {
	HitCount      uint64 `json:"hit_count"`
	MissCount     uint64 `json:"miss_count"`
	EntryCount    uint64 `json:"entry_count"`
	EvictionCount uint64 `json:"eviction_count"`
}

// KMSBackendCapabilities is the capability matrix of the active backend.
type KMSBackendCapabilities struct {
	Encrypt             bool `json:"encrypt"`
	Decrypt             bool `json:"decrypt"`
	GenerateDataKey     bool `json:"generate_data_key"`
	Rotate              bool `json:"rotate"`
	EnableDisable       bool `json:"enable_disable"`
	ScheduleDeletion    bool `json:"schedule_deletion"`
	Versioning          bool `json:"versioning"`
	PhysicalDelete      bool `json:"physical_delete"`
	UpdateKeyMetadata   bool `json:"update_key_metadata"`
	Rewrap              bool `json:"rewrap"`
	ProductionSupported bool `json:"production_supported"`
}

// KMSClusterConfigStatus is the cluster-wide configuration fingerprint view.
type KMSClusterConfigStatus struct {
	Consistent bool                  `json:"consistent"`
	Nodes      []KMSNodeConfigStatus `json:"nodes"`
}

// KMSNodeConfigStatus is a single node's configuration fingerprint.
type KMSNodeConfigStatus struct {
	Host              string  `json:"host"`
	ConfigFingerprint *string `json:"config_fingerprint"`
	Error             *string `json:"error"`
}

// KMSConfig is the response of GET /rustfs/admin/v3/kms/config.
type KMSConfig struct {
	Backend         string  `json:"backend"`
	CacheEnabled    bool    `json:"cache_enabled"`
	CacheMaxKeys    uint64  `json:"cache_max_keys"`
	CacheTTLSeconds uint64  `json:"cache_ttl_seconds"`
	DefaultKeyID    *string `json:"default_key_id"`
}

// KmsStatus returns the parsed KMS status document.
func (c *RustfsAdmin) KmsStatus() (KMSStatus, error) {
	var status KMSStatus
	err := c.getKmsJSON("kms/status", &status)
	return status, err
}

// KmsConfig returns the parsed KMS backend configuration document.
func (c *RustfsAdmin) KmsConfig() (KMSConfig, error) {
	var config KMSConfig
	err := c.getKmsJSON("kms/config", &config)
	return config, err
}

// getKmsJSON performs a GET against a KMS admin endpoint and decodes the JSON
// response body into out.
func (c *RustfsAdmin) getKmsJSON(relPath string, out interface{}) error {
	reqData := RequestData{
		Method:  "GET",
		RelPath: relPath,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
