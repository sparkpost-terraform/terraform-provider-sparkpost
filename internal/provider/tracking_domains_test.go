package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTrackingDomain_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"domain": "track.example.com",
				"secure": true,
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	got, err := client.GetTrackingDomain("track.example.com", 0)
	if err != nil {
		t.Fatalf("GetTrackingDomain() error = %v", err)
	}

	want := &TrackingDomain{Domain: "track.example.com", HTTPS: true}
	if *got != *want {
		t.Errorf("GetTrackingDomain() = %+v, want %+v", *got, *want)
	}
}

func TestGetTrackingDomain_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	_, err := client.GetTrackingDomain("track.example.com", 0)
	if !errors.Is(err, ErrTrackingDomainNotFound) {
		t.Errorf("GetTrackingDomain() error = %v, want ErrTrackingDomainNotFound", err)
	}
}

func TestDeleteTrackingDomain_NotFoundIsIdempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.DeleteTrackingDomain("track.example.com", 0); err != nil {
		t.Errorf("DeleteTrackingDomain() on an already-deleted domain should not error, got %v", err)
	}
}

func TestUpdateTrackingDomain_SendsSecureFlag(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.UpdateTrackingDomain("track.example.com", true, 0); err != nil {
		t.Fatalf("UpdateTrackingDomain() error = %v", err)
	}
	if secure, _ := gotBody["secure"].(bool); !secure {
		t.Errorf("request body secure = %v, want true", gotBody["secure"])
	}
}

func TestVerifyTrackingDomain_Unverified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{"verified": false, "cname_status": "pending"},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.VerifyTrackingDomain("track.example.com", 0); err == nil {
		t.Error("VerifyTrackingDomain() expected an error when unverified, got nil")
	}
}
