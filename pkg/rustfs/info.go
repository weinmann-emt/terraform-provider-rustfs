package rustfs

import (
	"context"
	"encoding/json"
)

// ServerInfo is the response of the GET /v3/info endpoint.
type ServerInfo struct {
	AdminDiscovery map[string]string `json:"admin_discovery"`
	BitrotSelftest string            `json:"bitrotSelftest"`
	Info           ClusterInfo       `json:"info"`
}

// ClusterInfo holds the cluster level information.
type ClusterInfo struct {
	Backend       BackendInfo                       `json:"backend"`
	Buckets       CountInfo                         `json:"buckets"`
	Deletemarkers CountInfo                         `json:"deletemarkers"`
	DeploymentID  string                            `json:"deploymentID"`
	Mode          string                            `json:"mode"`
	Objects       CountInfo                         `json:"objects"`
	Pools         map[string]map[string]PoolSetInfo `json:"pools"`
	Region        *string                           `json:"region"`
	Servers       []ServerEntry                     `json:"servers"`
	Usage         UsageInfo                         `json:"usage"`
	Versions      CountInfo                         `json:"versions"`
}

// BackendInfo describes the erasure / parity configuration of the cluster.
type BackendInfo struct {
	BackendType       string  `json:"backendType"`
	OfflineDisks      int64   `json:"offlineDisks"`
	OnlineDisks       int64   `json:"onlineDisks"`
	RRSCParity        int64   `json:"rrSCParity"`
	StandardSCParity  int64   `json:"standardSCParity"`
	TotalDrivesPerSet []int64 `json:"totalDrivesPerSet"`
	TotalSets         []int64 `json:"totalSets"`
	UnknownDisks      int64   `json:"unknownDisks"`
}

// CountInfo is a counter with an optional error field.
type CountInfo struct {
	Count int64 `json:"count"`
}

// UsageInfo describes aggregate usage.
type UsageInfo struct {
	Size int64 `json:"size"`
}

// ServerEntry describes a single server node in the cluster.
type ServerEntry struct {
	CommitID string      `json:"commitID"`
	Drives   []DriveInfo `json:"drives"`
	Endpoint string      `json:"endpoint"`
	MaxProcs int         `json:"max_procs"`
	MemStats MemStats    `json:"mem_stats"`
	NumCPU   int         `json:"num_cpu"`
	State    string      `json:"state"`
	Uptime   int64       `json:"uptime"`
	Version  string      `json:"version"`
}

// DriveInfo describes a single drive attached to a server node.
type DriveInfo struct {
	Availspace   int64   `json:"availspace"`
	Endpoint     string  `json:"endpoint"`
	Healing      bool    `json:"healing"`
	Local        bool    `json:"local"`
	Path         string  `json:"path"`
	RuntimeState string  `json:"runtimeState"`
	State        string  `json:"state"`
	Totalspace   int64   `json:"totalspace"`
	Usedspace    int64   `json:"usedspace"`
	UUID         string  `json:"uuid"`
	Utilization  float64 `json:"utilization"`
}

// MemStats describes the Go runtime memory statistics of a server node.
type MemStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	HeapAlloc  uint64 `json:"heap_alloc"`
}

// PoolSetInfo describes a single erasure set within a storage pool.
type PoolSetInfo struct {
	DeleteMarkersCount int64 `json:"deleteMarkersCount"`
	HealDisks          int64 `json:"healDisks"`
	ID                 int64 `json:"id"`
	ObjectsCount       int64 `json:"objectsCount"`
	RawCapacity        int64 `json:"rawCapacity"`
	RawUsage           int64 `json:"rawUsage"`
	Usage              int64 `json:"usage"`
	VersionsCount      int64 `json:"versionsCount"`
}

// ServerInfo returns cluster and server information from GET /v3/info.
func (c *RustfsAdmin) ServerInfo() (ServerInfo, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "info",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return ServerInfo{}, err
	}
	defer resp.Body.Close()
	var info ServerInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}
