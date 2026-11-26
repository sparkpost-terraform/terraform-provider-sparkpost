package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateDomain_SetsSubaccountHeader(t *testing.T) {
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-MSYS-SUBACCOUNT")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.CreateDomain("example.com", 12345, false); err != nil {
		t.Fatalf("CreateDomain() error = %v", err)
	}
	if gotHeader != "12345" {
		t.Errorf("X-MSYS-SUBACCOUNT header = %q, want %q", gotHeader, "12345")
	}
}

func TestGetDomain_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"shared_with_subaccounts":  true,
				"is_default_bounce_domain": false,
				"dkim": map[string]interface{}{
					"selector": "scph0118",
					"public":   "MIGfMA0GCSqGSIb3DQEB",
					"headers":  "from:to:subject:date",
				},
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	got, err := client.GetDomain("example.com", 0)
	if err != nil {
		t.Fatalf("GetDomain() error = %v", err)
	}

	want := &TargetDomain{Domain: "example.com", SharedWithSubaccounts: true, DefaultBounceDomain: false, DKIMSelector: "scph0118", DKIMPublicKey: "MIGfMA0GCSqGSIb3DQEB"}
	if *got != *want {
		t.Errorf("GetDomain() = %+v, want %+v", *got, *want)
	}
}

func TestGetDomain_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	_, err := client.GetDomain("example.com", 0)
	if !errors.Is(err, ErrDomainNotFound) {
		t.Errorf("GetDomain() error = %v, want ErrDomainNotFound", err)
	}
}

func TestDeleteDomain_NotFoundIsIdempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.DeleteDomain("example.com", 0); err != nil {
		t.Errorf("DeleteDomain() on an already-deleted domain should not error, got %v", err)
	}
}

func TestDeleteDomain_OtherErrorPropagates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.DeleteDomain("example.com", 0); err == nil {
		t.Error("DeleteDomain() expected an error for a 500 response, got nil")
	}
}

func TestVerifyDomainOwnership_Unverified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{"ownership_verified": false},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.VerifyDomainOwnership("example.com", 0); err == nil {
		t.Error("VerifyDomainOwnership() expected an error when unverified, got nil")
	}
}

func TestGetTrackingDomainAssociation_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	_, err := client.GetTrackingDomainAssociation("example.com", 0, "track.example.com")
	if !errors.Is(err, ErrDomainNotFound) {
		t.Errorf("GetTrackingDomainAssociation() error = %v, want ErrDomainNotFound", err)
	}
}

func TestCreateDomain_LeavesDefaultBounceOut(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.CreateDomain("bounces.example.com", 0, false); err != nil {
		t.Fatalf("CreateDomain() error = %v", err)
	}
	if _, ok := gotBody["is_default_bounce_domain"]; ok {
		t.Errorf("CreateDomain() body = %v, want no is_default_bounce_domain", gotBody)
	}
}

func TestSetDefaultBounceDomain_Success(t *testing.T) {
	var gotMethod, gotPath, gotHeader string
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotHeader = r.Method, r.URL.Path, r.Header.Get("X-MSYS-SUBACCOUNT")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.SetDefaultBounceDomain("bounces.example.com", 12345, true); err != nil {
		t.Fatalf("SetDefaultBounceDomain() error = %v", err)
	}
	if gotMethod != http.MethodPut || gotPath != "/sending-domains/bounces.example.com" {
		t.Errorf("request = %s %s, want PUT /sending-domains/bounces.example.com", gotMethod, gotPath)
	}
	if gotHeader != "12345" {
		t.Errorf("X-MSYS-SUBACCOUNT header = %q, want %q", gotHeader, "12345")
	}
	if gotBody["is_default_bounce_domain"] != true {
		t.Errorf("body = %v, want is_default_bounce_domain true", gotBody)
	}
}

func TestSetDefaultBounceDomain_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	err := client.SetDefaultBounceDomain("bounces.example.com", 0, false)
	if !errors.Is(err, ErrDomainNotFound) {
		t.Errorf("SetDefaultBounceDomain() error = %v, want ErrDomainNotFound", err)
	}
}
