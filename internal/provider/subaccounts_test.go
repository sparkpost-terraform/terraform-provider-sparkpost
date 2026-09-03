package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSubaccounts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 1, "name": "primary"},
				{"id": 2, "name": "secondary"},
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	got, err := client.ListSubaccounts()
	if err != nil {
		t.Fatalf("ListSubaccounts() error = %v", err)
	}
	if len(got) != 2 || got[0].Name != "primary" || got[1].ID != 2 {
		t.Errorf("ListSubaccounts() = %+v, want [{1 primary} {2 secondary}]", got)
	}
}

func TestListSubaccounts_RequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if _, err := client.ListSubaccounts(); err == nil {
		t.Error("ListSubaccounts() expected an error for a 500 response, got nil")
	}
}

func TestCreateSubaccount_Success(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"subaccount_id": 888,
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	id, err := client.CreateSubaccount("Sparkle Ponies", "my_pool")
	if err != nil {
		t.Fatalf("CreateSubaccount() error = %v", err)
	}
	if id != 888 {
		t.Errorf("CreateSubaccount() id = %d, want 888", id)
	}
	if gotBody["setup_api_key"] != false {
		t.Errorf("CreateSubaccount() setup_api_key = %v, want false", gotBody["setup_api_key"])
	}
	if gotBody["ip_pool"] != "my_pool" {
		t.Errorf("CreateSubaccount() ip_pool = %v, want my_pool", gotBody["ip_pool"])
	}
}

func TestGetSubaccount_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"id":      123,
				"name":    "Joes Garage",
				"status":  "active",
				"ip_pool": "assigned_ip_pool",
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	got, err := client.GetSubaccount(123)
	if err != nil {
		t.Fatalf("GetSubaccount() error = %v", err)
	}

	want := &Subaccount{ID: 123, Name: "Joes Garage", Status: "active", IPPool: "assigned_ip_pool"}
	if *got != *want {
		t.Errorf("GetSubaccount() = %+v, want %+v", *got, *want)
	}
}

func TestGetSubaccount_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	_, err := client.GetSubaccount(123)
	if !errors.Is(err, ErrSubaccountNotFound) {
		t.Errorf("GetSubaccount() error = %v, want ErrSubaccountNotFound", err)
	}
}

func TestUpdateSubaccount_Success(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"message": "Successfully updated subaccount information",
			},
		})
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	if err := client.UpdateSubaccount(123, "Renamed", "terminated", ""); err != nil {
		t.Fatalf("UpdateSubaccount() error = %v", err)
	}
	if gotBody["status"] != "terminated" {
		t.Errorf("UpdateSubaccount() status = %v, want terminated", gotBody["status"])
	}
}

func TestUpdateSubaccount_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSparkPostClient(server.URL+"/", "test-key")
	err := client.UpdateSubaccount(123, "Renamed", "active", "")
	if !errors.Is(err, ErrSubaccountNotFound) {
		t.Errorf("UpdateSubaccount() error = %v, want ErrSubaccountNotFound", err)
	}
}
