package rustfs

import (
	"context"
	"encoding/json"
)

// StorageInfo holds the decoded response from GET /rustfs/admin/v3/storageinfo.
type StorageInfo struct {
	// AdminDiscovery lists related admin API endpoints.
	AdminDiscovery *AdminDiscovery `json:"admin_discovery,omitempty"`
	// Info contains the storage layout and health breakdown.
	Info *StorageInfoDetails `json:"info,omitempty"`
}

// AdminDiscovery lists related admin API endpoints.
type AdminDiscovery struct {
	ClusterSnapshot     string `json:"clusterSnapshot,omitempty"`
	ExtensionsCatalog   string `json:"extensionsCatalog,omitempty"`
	RuntimeCapabilities string `json:"runtimeCapabilities,omitempty"`
}

// StorageInfoDetails holds the backend and per-drive breakdown.
type StorageInfoDetails struct {
	Backend *BackendInfo `json:"backend,omitempty"`
	Disks   []DiskInfo   `json:"disks,omitempty"`
}

// BackendInfo describes the storage backend layout.
type BackendInfo struct {
	BackendType        string `json:"BackendType,omitempty"`
	DrivesPerSet       []int  `json:"DrivesPerSet,omitempty"`
	OfflineDisks       any    `json:"OfflineDisks,omitempty"`
	OnlineDisks        any    `json:"OnlineDisks,omitempty"`
	RRSCData           []int  `json:"RRSCData,omitempty"`
	RRSCParities       []int  `json:"RRSCParities,omitempty"`
	RRSCParity         int    `json:"RRSCParity,omitempty"`
	StandardSCData     []int  `json:"StandardSCData,omitempty"`
	StandardSCParities []int  `json:"StandardSCParities,omitempty"`
	StandardSCParity   int    `json:"StandardSCParity,omitempty"`
	TotalSets          []int  `json:"TotalSets,omitempty"`
}

// DiskInfo describes a single drive.
type DiskInfo struct {
	DiskIndex    int    `json:"disk_index,omitempty"`
	Endpoint     string `json:"endpoint,omitempty"`
	AvailSpace   int64  `json:"availspace,omitempty"`
	FreeInodes   int64  `json:"free_inodes,omitempty"`
	Healing      bool   `json:"healing,omitempty"`
	Local        bool   `json:"local,omitempty"`
	Path         string `json:"path,omitempty"`
	PoolIndex    int    `json:"pool_index,omitempty"`
	RootDisk     bool   `json:"rootDisk,omitempty"`
	RuntimeState string `json:"runtimeState,omitempty"`
	Scanning     bool   `json:"scanning,omitempty"`
	SetIndex     int    `json:"set_index,omitempty"`
	State        string `json:"state,omitempty"`
	TotalSpace   int64  `json:"totalspace,omitempty"`
	UsedInodes   int64  `json:"used_inodes,omitempty"`
	UsedSpace    int64  `json:"usedspace,omitempty"`
	UUID         string `json:"uuid,omitempty"`
	Major        int    `json:"major,omitempty"`
	Minor        int    `json:"minor,omitempty"`
}

// StorageInfo returns storage layout and health from GET /v3/storageinfo.
func (c *RustfsAdmin) StorageInfo() (*StorageInfo, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "storageinfo",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info StorageInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
