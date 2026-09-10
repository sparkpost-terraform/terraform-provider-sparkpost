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

func TestCheckTrackingDomainCertificateEligibility_Eligible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"domain":                     "track.example.com",
				"supportsManagedCertificate": true,
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	eligible, err := client.CheckTrackingDomainCertificateEligibility("track.example.com", 0)
	if err != nil {
		t.Fatalf("CheckTrackingDomainCertificateEligibility() error = %v", err)
	}
	if !eligible {
		t.Error("CheckTrackingDomainCertificateEligibility() = false, want true")
	}
}

func TestCheckTrackingDomainCertificateEligibility_Ineligible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"domain":                     "track.example.com",
				"supportsManagedCertificate": false,
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	eligible, err := client.CheckTrackingDomainCertificateEligibility("track.example.com", 0)
	if err != nil {
		t.Fatalf("CheckTrackingDomainCertificateEligibility() error = %v", err)
	}
	if eligible {
		t.Error("CheckTrackingDomainCertificateEligibility() = true, want false")
	}
}

func TestEnableTrackingDomainManagedCertificate_Success(t *testing.T) {
	var gotSubaccountHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSubaccountHeader = r.Header.Get("X-MSYS-SUBACCOUNT")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{"message": "Certificate issuance initiated"},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.EnableTrackingDomainManagedCertificate("track.example.com", 42); err != nil {
		t.Fatalf("EnableTrackingDomainManagedCertificate() error = %v", err)
	}
	if gotSubaccountHeader != "42" {
		t.Errorf("X-MSYS-SUBACCOUNT header = %q, want %q", gotSubaccountHeader, "42")
	}
}

func TestEnableTrackingDomainManagedCertificate_NotEligible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Domain is not eligible for managed certificates due to Let's Encrypt policies"},
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	err := client.EnableTrackingDomainManagedCertificate("track.example.com", 0)
	if err == nil {
		t.Fatal("EnableTrackingDomainManagedCertificate() expected an error, got nil")
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
