package rustfs

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExportIam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-zip-data"))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	data, err := client.ExportIam()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "fake-zip-data" {
		t.Errorf("expected fake-zip-data, got %s", string(data))
	}
}

func TestImportIam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	err := client.ImportIam([]byte("fake-zip-data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
