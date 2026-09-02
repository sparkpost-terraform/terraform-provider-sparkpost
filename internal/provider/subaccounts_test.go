package provider

import (
	"encoding/json"
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
