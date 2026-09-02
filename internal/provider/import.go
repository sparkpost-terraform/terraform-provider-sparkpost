package provider

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// parseImportID splits a "domain" or "domain,subaccount" import identifier,
// as used by resources keyed on the pair (domain, subaccount).
func parseImportID(id string) (domain string, subaccount int64, hasSubaccount bool, diags diag.Diagnostics) {
	parts := strings.SplitN(id, ",", 2)

	domain = strings.TrimSpace(parts[0])
	if domain == "" {
		diags.AddError(
			"Invalid Import ID",
			`Expected import ID in the format "domain" or "domain,subaccount", got: `+id,
		)
		return "", 0, false, diags
	}

	if len(parts) == 1 {
		return domain, 0, false, diags
	}

	sub, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		diags.AddError(
			"Invalid Import ID",
			`Expected import ID in the format "domain" or "domain,subaccount", subaccount must be an integer, got: `+id,
		)
		return "", 0, false, diags
	}

	return domain, sub, true, diags
}
