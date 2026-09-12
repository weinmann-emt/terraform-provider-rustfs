package rustfs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func sampleHealthInfoJSON() string {
	return `{
		"version": "refs/tags/1.0.0-rc.5",
		"deployment_id": "3d6a83f0-8b13-49e0-bd0e-3a921f256a6c",
		"region": "us-east-1",
		"timestamp": "2026-09-08T12:20:11.301344385Z",
		"cpu": {"logical_cores": 6, "brand": "", "frequency_mhz": 48, "usage_percent": 1.5873015721638997},
		"memory": {"total_bytes": 8289050624, "used_bytes": 832872448, "available_bytes": 7456178176, "total_swap_bytes": 0, "used_swap_bytes": 0},
		"os": {"os": "linux", "kernel_version": "7.1.3-200.fc44.aarch64", "os_version": "Linux (Alpine Linux 3.24.1)", "hostname": "30592b34dfea", "arch": "aarch64", "uptime_secs": 88421},
		"process": {"pid": 1, "cpu_usage_percent": 0.0, "memory_bytes": 176513024},
		"drives": [{"endpoint": "/data", "drive_path": "/data", "state": "ok", "total_space": 63728975872, "used_space": 4699365376, "available_space": 59029610496, "read_throughput": 0.0, "write_throughput": 0.0, "read_latency": 0.0, "write_latency": 0.0}],
		"unsupported_probes": ["perf-net", "perf-drive-obd", "config-obd", "sys-services"]
	}`
}

func sampleObdInfoJSON() string {
	return `{
		"version": "refs/tags/1.0.0-rc.5",
		"deployment_id": "3d6a83f0-8b13-49e0-bd0e-3a921f256a6c",
		"region": "us-east-1",
		"timestamp": "2026-09-08T12:20:11.5129739Z",
		"cpu": {"logical_cores": 6, "brand": "", "frequency_mhz": 48, "usage_percent": 0.7936507860819498},
		"memory": {"total_bytes": 8289050624, "used_bytes": 832872448, "available_bytes": 7456178176, "total_swap_bytes": 0, "used_swap_bytes": 0},
		"os": {"os": "linux", "kernel_version": "7.1.3-200.fc44.aarch64", "os_version": "Linux (Alpine Linux 3.24.1)", "hostname": "30592b34dfea", "arch": "aarch64", "uptime_secs": 88421},
		"process": {"pid": 1, "cpu_usage_percent": 0.0, "memory_bytes": 176533504},
		"drives": [{"endpoint": "/data", "drive_path": "/data", "state": "degraded", "total_space": 63728975872, "used_space": 4699365376, "available_space": 59029610496, "read_throughput": 0.0, "write_throughput": 0.0, "read_latency": 0.0, "write_latency": 0.0}],
		"unsupported_probes": ["perf-net", "perf-drive-obd", "config-obd", "sys-services"]
	}`
}

func newHealthTestClient(t *testing.T, handler http.HandlerFunc) *RustfsAdmin {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := New(&RustfsAdminConfig{
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
		AccessKey:    "admin",
		AccessSecret: "secret",
	})
	return &client
}

func TestGetHealthInfo(t *testing.T) {
	client := newHealthTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rustfs/admin/v3/healthinfo") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleHealthInfoJSON()))
	})

	info, err := client.GetHealthInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "refs/tags/1.0.0-rc.5" {
		t.Errorf("expected version, got %q", info.Version)
	}
	if info.Region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %q", info.Region)
	}
	if len(info.Drives) != 1 {
		t.Fatalf("expected 1 drive, got %d", len(info.Drives))
	}
	if info.Drives[0].State != "ok" {
		t.Errorf("expected drive state ok, got %q", info.Drives[0].State)
	}
	if info.CPU.LogicalCores != 6 {
		t.Errorf("expected 6 logical cores, got %d", info.CPU.LogicalCores)
	}
	if info.OS.Arch != "aarch64" {
		t.Errorf("expected arch aarch64, got %q", info.OS.Arch)
	}
	if len(info.UnsupportedProbes) != 4 {
		t.Errorf("expected 4 unsupported probes, got %d", len(info.UnsupportedProbes))
	}
}

func TestGetHealthInfoDecodesDegradedDrive(t *testing.T) {
	client := newHealthTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleObdInfoJSON()))
	})

	info, err := client.GetHealthInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Drives[0].State != "degraded" {
		t.Errorf("expected degraded drive state, got %q", info.Drives[0].State)
	}
}

func TestGetObdInfo(t *testing.T) {
	client := newHealthTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rustfs/admin/v3/obdinfo") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleObdInfoJSON()))
	})

	info, err := client.GetObdInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Drives[0].State != "degraded" {
		t.Errorf("expected degraded drive state, got %q", info.Drives[0].State)
	}
}

func TestGetHealthInfoUnmarshalRoundTrip(t *testing.T) {
	client := newHealthTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleHealthInfoJSON()))
	})

	info, err := client.GetHealthInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	raw, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	var back map[string]interface{}
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if _, ok := back["drives"]; !ok {
		t.Errorf("expected drives key in re-marshalled output")
	}
}
