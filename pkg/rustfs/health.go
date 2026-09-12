package rustfs

import (
	"context"
	"encoding/json"
)

// HealthInfo models the JSON response returned by both
// GET /v3/healthinfo and GET /v3/obdinfo.
type HealthInfo struct {
	Version           string        `json:"version"`
	DeploymentID      string        `json:"deployment_id"`
	Region            string        `json:"region"`
	Timestamp         string        `json:"timestamp"`
	CPU               HealthCPU     `json:"cpu"`
	Memory            HealthMemory  `json:"memory"`
	OS                HealthOS      `json:"os"`
	Process           HealthProcess `json:"process"`
	Drives            []HealthDrive `json:"drives"`
	UnsupportedProbes []string      `json:"unsupported_probes"`
}

// HealthCPU describes the logical CPU metrics reported by the health info endpoints.
type HealthCPU struct {
	LogicalCores int     `json:"logical_cores"`
	Brand        string  `json:"brand"`
	FrequencyMHz float64 `json:"frequency_mhz"`
	UsagePercent float64 `json:"usage_percent"`
}

// HealthMemory describes the host memory metrics reported by the health info endpoints.
type HealthMemory struct {
	TotalBytes     int64 `json:"total_bytes"`
	UsedBytes      int64 `json:"used_bytes"`
	AvailableBytes int64 `json:"available_bytes"`
	TotalSwapBytes int64 `json:"total_swap_bytes"`
	UsedSwapBytes  int64 `json:"used_swap_bytes"`
}

// HealthOS describes the host operating system reported by the health info endpoints.
type HealthOS struct {
	OS            string `json:"os"`
	KernelVersion string `json:"kernel_version"`
	OSVersion     string `json:"os_version"`
	Hostname      string `json:"hostname"`
	Arch          string `json:"arch"`
	UptimeSecs    int64  `json:"uptime_secs"`
}

// HealthProcess describes the RustFS process metrics reported by the health info endpoints.
type HealthProcess struct {
	PID             int     `json:"pid"`
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
	MemoryBytes     int64   `json:"memory_bytes"`
}

// HealthDrive describes a single drive state reported by the health info endpoints.
type HealthDrive struct {
	Endpoint        string  `json:"endpoint"`
	DrivePath       string  `json:"drive_path"`
	State           string  `json:"state"`
	TotalSpace      int64   `json:"total_space"`
	UsedSpace       int64   `json:"used_space"`
	AvailableSpace  int64   `json:"available_space"`
	ReadThroughput  float64 `json:"read_throughput"`
	WriteThroughput float64 `json:"write_throughput"`
	ReadLatency     float64 `json:"read_latency"`
	WriteLatency    float64 `json:"write_latency"`
}

func (c *RustfsAdmin) getHealthInfo(relPath string) (*HealthInfo, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: relPath,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var info HealthInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetHealthInfo returns cluster health from GET /v3/healthinfo.
func (c *RustfsAdmin) GetHealthInfo() (*HealthInfo, error) {
	return c.getHealthInfo("healthinfo")
}

// GetObdInfo returns OBD diagnostics from GET /v3/obdinfo.
func (c *RustfsAdmin) GetObdInfo() (*HealthInfo, error) {
	return c.getHealthInfo("obdinfo")
}
