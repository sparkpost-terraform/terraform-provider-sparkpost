package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	req, err := client.newRequest("GET", "ping", nil)
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}

	resp, err := client.doRequest(req, http.StatusOK)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
}

func TestDoRequest_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	req, err := client.newRequest("GET", "missing", nil)
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}

	_, err = client.doRequest(req, http.StatusOK)
	if err == nil {
		t.Fatal("doRequest() expected an error, got nil")
	}
	if !isNotFound(err) {
		t.Fatalf("isNotFound(%v) = false, want true", err)
	}
}

func TestIsNotFound(t *testing.T) {
	if isNotFound(nil) {
		t.Error("isNotFound(nil) = true, want false")
	}
	if isNotFound(&apiError{StatusCode: http.StatusInternalServerError, Status: "500 Internal Server Error"}) {
		t.Error("isNotFound(500) = true, want false")
	}
	if !isNotFound(&apiError{StatusCode: http.StatusNotFound, Status: "404 Not Found"}) {
		t.Error("isNotFound(404) = false, want true")
	}
}

func TestNewRequest_SetsAuthAndContentType(t *testing.T) {
	client := NewSparkPostClient("https://example.invalid/", "test-key")
	req, err := client.newRequest("POST", "sending-domains", map[string]interface{}{"domain": "example.com"})
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}

	if got := req.Header.Get("Authorization"); got != "test-key" {
		t.Errorf("Authorization header = %q, want %q", got, "test-key")
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type header = %q, want %q", got, "application/json")
	}
	if got := req.URL.String(); got != "https://example.invalid/sending-domains" {
		t.Errorf("request URL = %q, want %q", got, "https://example.invalid/sending-domains")
	}
}
