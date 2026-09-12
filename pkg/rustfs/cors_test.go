package rustfs

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestCorsServer(t *testing.T, handler http.HandlerFunc) RustfsAdmin {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"
	return client
}

func TestSetBucketCorsConfiguration(t *testing.T) {
	client := newTestCorsServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/test-bucket" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		if _, ok := r.URL.Query()["cors"]; !ok {
			t.Errorf("expected cors query param, got %s", r.URL.RawQuery)
		}
		body, _ := io.ReadAll(r.Body)
		for _, want := range []string{"CORSConfiguration", "<AllowedMethod>GET</AllowedMethod>", "<AllowedOrigin>https://example.com</AllowedOrigin>", "<MaxAgeSeconds>3000</MaxAgeSeconds>"} {
			if !strings.Contains(string(body), want) {
				t.Errorf("expected body to contain %q, got %s", want, string(body))
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	config := &CORSConfiguration{
		Rules: []CORSRule{
			{
				ID:             "corsrule1",
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"*"},
				ExposeHeaders:  []string{"ETag"},
				MaxAgeSeconds:  3000,
			},
		},
	}

	if err := client.SetBucketCorsConfiguration("test-bucket", config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBucketCorsConfiguration(t *testing.T) {
	client := newTestCorsServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/test-bucket" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		if _, ok := r.URL.Query()["cors"]; !ok {
			t.Errorf("expected cors query param, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<CORSConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><CORSRule><ID>corsrule1</ID><AllowedHeader>*</AllowedHeader><AllowedMethod>GET</AllowedMethod><AllowedMethod>PUT</AllowedMethod><AllowedOrigin>https://example.com</AllowedOrigin><ExposeHeader>ETag</ExposeHeader><MaxAgeSeconds>3000</MaxAgeSeconds></CORSRule></CORSConfiguration>`))
	})

	config, err := client.GetBucketCorsConfiguration("test-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(config.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(config.Rules))
	}
	rule := config.Rules[0]
	if rule.ID != "corsrule1" {
		t.Errorf("expected id corsrule1, got %s", rule.ID)
	}
	if rule.MaxAgeSeconds != 3000 {
		t.Errorf("expected max age 3000, got %d", rule.MaxAgeSeconds)
	}
	if len(rule.AllowedMethods) != 2 || rule.AllowedMethods[0] != "GET" || rule.AllowedMethods[1] != "PUT" {
		t.Errorf("unexpected allowed methods: %v", rule.AllowedMethods)
	}
	if len(rule.AllowedOrigins) != 1 || rule.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("unexpected allowed origins: %v", rule.AllowedOrigins)
	}
	if len(rule.AllowedHeaders) != 1 || rule.AllowedHeaders[0] != "*" {
		t.Errorf("unexpected allowed headers: %v", rule.AllowedHeaders)
	}
	if len(rule.ExposeHeaders) != 1 || rule.ExposeHeaders[0] != "ETag" {
		t.Errorf("unexpected expose headers: %v", rule.ExposeHeaders)
	}
}

func TestGetBucketCorsConfigurationNotFound(t *testing.T) {
	client := newTestCorsServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchCORSConfiguration</Code><Message>The CORS configuration does not exist</Message></Error>`))
	})

	_, err := client.GetBucketCorsConfiguration("test-bucket")
	if err == nil {
		t.Fatal("expected error for missing CORS configuration")
	}
	if !strings.Contains(err.Error(), "NoSuchCORSConfiguration") {
		t.Errorf("expected NoSuchCORSConfiguration in error, got %v", err)
	}
}

func TestDeleteBucketCorsConfiguration(t *testing.T) {
	client := newTestCorsServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/test-bucket" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		if _, ok := r.URL.Query()["cors"]; !ok {
			t.Errorf("expected cors query param, got %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.DeleteBucketCorsConfiguration("test-bucket"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCORSConfigurationRoundTrip(t *testing.T) {
	config := &CORSConfiguration{
		Rules: []CORSRule{
			{
				ID:             "corsrule1",
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "PUT"},
				AllowedHeaders: []string{"*"},
				ExposeHeaders:  []string{"ETag"},
				MaxAgeSeconds:  3000,
			},
		},
	}

	data, err := xml.Marshal(config)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded CORSConfiguration
	if err := xml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(decoded.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(decoded.Rules))
	}
	rule := decoded.Rules[0]
	if rule.ID != config.Rules[0].ID || rule.MaxAgeSeconds != config.Rules[0].MaxAgeSeconds {
		t.Errorf("round trip mismatch: %+v", rule)
	}
}
