package provider

import "testing"

func TestParseImportID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantDomain string
		wantSub    int64
		wantHasSub bool
		wantErr    bool
	}{
		{name: "domain only", id: "example.com", wantDomain: "example.com", wantHasSub: false},
		{name: "domain and subaccount", id: "example.com,12345", wantDomain: "example.com", wantSub: 12345, wantHasSub: true},
		{name: "whitespace is trimmed", id: " example.com , 12345 ", wantDomain: "example.com", wantSub: 12345, wantHasSub: true},
		{name: "empty domain", id: ",12345", wantErr: true},
		{name: "non-numeric subaccount", id: "example.com,not-a-number", wantErr: true},
		{name: "empty id", id: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, sub, hasSub, diags := parseImportID(tt.id)

			if tt.wantErr {
				if !diags.HasError() {
					t.Fatalf("parseImportID(%q) expected an error, got none", tt.id)
				}
				return
			}

			if diags.HasError() {
				t.Fatalf("parseImportID(%q) unexpected error: %v", tt.id, diags)
			}
			if domain != tt.wantDomain {
				t.Errorf("domain = %q, want %q", domain, tt.wantDomain)
			}
			if sub != tt.wantSub {
				t.Errorf("subaccount = %d, want %d", sub, tt.wantSub)
			}
			if hasSub != tt.wantHasSub {
				t.Errorf("hasSubaccount = %v, want %v", hasSub, tt.wantHasSub)
			}
		})
	}
}
