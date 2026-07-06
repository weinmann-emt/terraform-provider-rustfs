package rustfs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequest_NetworkError(t *testing.T) {
	client := New(&RustfsAdminConfig{
		Endpoint:  "127.0.0.1:1",
		AccessKey: "test",
	})

	reqData := RequestData{
		Method:  "GET",
		RelPath: "test",
	}

	resp, err := client.doRequest(context.Background(), reqData)
	if err == nil {
		t.Fatal("expected error from network failure, got nil")
	}
	if resp != nil {
		t.Fatal("expected nil response on network failure")
	}
}

func TestDoRequest_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "test",
	})
	client.accessSecret = "test"

	reqData := RequestData{
		Method:  "GET",
		RelPath: "/",
	}

	resp, err := client.doRequest(context.Background(), reqData)
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
	if resp == nil {
		t.Fatal("expected non-nil response for HTTP error")
	}
}

func TestDoDirectRequest_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "test",
	})
	client.accessSecret = "test"

	reqData := RequestData{
		Method:  "GET",
		RelPath: "/",
	}

	resp, err := client.DoDirectRequest(context.Background(), reqData)
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
	if resp == nil {
		t.Fatal("expected non-nil response for HTTP error")
	}
}
