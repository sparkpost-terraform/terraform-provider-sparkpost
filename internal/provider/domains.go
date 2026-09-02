package provider

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// TargetDomain represents a sending domain. The SparkPost API does not echo
// the domain name back in GET responses, so callers must supply it.
type TargetDomain struct {
	Domain                string `json:"domain"`
	SharedWithSubaccounts bool   `json:"shared_with_subaccounts"`
	DefaultBounceDomain   bool   `json:"is_default_bounce_domain"`
}

func (c *SparkPostClient) CreateDomain(domain string, subaccount int, shared bool, defaultBounce bool) error {
	body := map[string]interface{}{
		"domain":                   domain,
		"shared_with_subaccounts":  shared,
		"is_default_bounce_domain": defaultBounce,
	}

	req, err := c.newRequest("POST", "sending-domains", body)
	if err != nil {
		return fmt.Errorf("failed to build create domain request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return fmt.Errorf("create domain request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

func (c *SparkPostClient) GetDomain(domain string, subaccount int) (*TargetDomain, error) {
	endpoint := fmt.Sprintf("sending-domains/%s", domain)

	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build get domain request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrDomainNotFound
		}
		return nil, fmt.Errorf("get domain request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Results struct {
			SharedWithSubaccounts bool `json:"shared_with_subaccounts"`
			IsDefaultBounceDomain bool `json:"is_default_bounce_domain"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse get domain response: %w", err)
	}

	return &TargetDomain{
		Domain:                domain,
		SharedWithSubaccounts: body.Results.SharedWithSubaccounts,
		DefaultBounceDomain:   body.Results.IsDefaultBounceDomain,
	}, nil
}

func (c *SparkPostClient) DeleteDomain(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("sending-domains/%s", domain)

	req, err := c.newRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to build delete domain request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 204)
	if err != nil {
		if isNotFound(err) {
			// Already gone: deleting a nonexistent domain is not an error.
			return nil
		}
		return fmt.Errorf("delete domain request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

func (c *SparkPostClient) VerifyDomainOwnership(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("sending-domains/%s/verify", domain)

	body := map[string]interface{}{
		"dkim_verify": true,
	}

	req, err := c.newRequest("POST", endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return fmt.Errorf("verification request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var respBody struct {
		Results struct {
			OwnershipVerified bool `json:"ownership_verified"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to parse verification response: %w", err)
	}

	if !respBody.Results.OwnershipVerified {
		return fmt.Errorf("verification failed: ownership_verified = '%t'", respBody.Results.OwnershipVerified)
	}

	return nil
}

func (c *SparkPostClient) VerifyDomainCNAME(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("sending-domains/%s/verify", domain)

	body := map[string]interface{}{
		"cname_verify": true,
	}

	req, err := c.newRequest("POST", endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return fmt.Errorf("verification request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var respBody struct {
		Results struct {
			CNAMEStatus string `json:"cname_status"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to parse verification response: %w", err)
	}

	if respBody.Results.CNAMEStatus != "valid" {
		return fmt.Errorf("verification failed: cname_status = '%s'", respBody.Results.CNAMEStatus)
	}

	return nil
}

func (c *SparkPostClient) AssociateTrackingDomain(domain string, subaccount int, trackingDomain string) error {
	endpoint := fmt.Sprintf("sending-domains/%s", domain)

	body := map[string]interface{}{
		"tracking_domain": trackingDomain,
	}

	req, err := c.newRequest("PUT", endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return fmt.Errorf("association request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var respBody struct {
		Results struct {
			Message string `json:"message"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to parse association response: %w", err)
	}

	if respBody.Results.Message != "Successfully Updated Domain." {
		return fmt.Errorf("association failed with '%s'", respBody.Results.Message)
	}

	return nil
}

func (c *SparkPostClient) GetTrackingDomainAssociation(domain string, subaccount int, trackingDomain string) (string, error) {
	endpoint := fmt.Sprintf("sending-domains/%s", domain)

	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		if isNotFound(err) {
			return "", ErrDomainNotFound
		}
		return "", fmt.Errorf("get tracking association request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var respBody struct {
		Results struct {
			TrackingDomain string `json:"tracking_domain"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("failed to parse get tracking domain association response: %w", err)
	}

	return respBody.Results.TrackingDomain, nil
}

var ErrDomainNotFound = fmt.Errorf("sending domain not found")
