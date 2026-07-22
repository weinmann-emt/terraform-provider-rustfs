package rustfs

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExportBucketMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-metadata-zip"))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	data, err := client.ExportBucketMetadata()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "fake-metadata-zip" {
		t.Errorf("expected fake-metadata-zip, got %s", string(data))
	}
}

func TestImportBucketMetadata(t *testing.T) {
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

	err := client.ImportBucketMetadata([]byte("fake-metadata-zip"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
