package provider

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type TrackingDomain struct {
	Domain                 string `json:"domain"`
	HTTPS                  bool   `json:"secure"`
	UsesManagedCertificate bool   `json:"uses_managed_certificate"`
}

// CreateTrackingDomain creates the tracking domain without HTTPS. Use
// UpdateTrackingDomain (via sparkpost_tracking_domain_https_configuration) to
// enable it afterwards - SparkPost's verify endpoint checks for a valid SSL
// certificate when secure=true, which only exists once a managed certificate
// has been enabled, which itself requires the domain to already be verified.
func (c *SparkPostClient) CreateTrackingDomain(domain string, subaccount int) error {
	body := map[string]interface{}{
		"domain": domain,
		"secure": false,
	}

	req, err := c.newRequest("POST", "tracking-domains", body)
	if err != nil {
		return fmt.Errorf("failed to build create tracking domain request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return fmt.Errorf("create tracking domain request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

func (c *SparkPostClient) GetTrackingDomain(domain string, subaccount int) (*TrackingDomain, error) {
	endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build get tracking domain request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrTrackingDomainNotFound
		}
		return nil, fmt.Errorf("get tracking domain request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Results TrackingDomain `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse get tracking domain response: %w", err)
	}

	return &body.Results, nil
}

func (c *SparkPostClient) DeleteTrackingDomain(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to build delete tracking domain request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 204)
	if err != nil {
		if isNotFound(err) {
			// Already gone: deleting a nonexistent tracking domain is not an error.
			return nil
		}
		return fmt.Errorf("delete tracking domain request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

func (c *SparkPostClient) UpdateTrackingDomain(domain string, https bool, subaccount int) error {
	body := map[string]interface{}{
		"secure": https,
	}

	endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("PUT", endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to build update tracking domain request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		if isNotFound(err) {
			return ErrTrackingDomainNotFound
		}
		return fmt.Errorf("update tracking domain request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

func (c *SparkPostClient) VerifyTrackingDomain(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("tracking-domains/%s/verify", domain)

	req, err := c.newRequest("POST", endpoint, nil)
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
			Verified    bool   `json:"verified"`
			CNAMEStatus string `json:"cname_status"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to parse verification response: %w", err)
	}

	if !respBody.Results.Verified {
		return fmt.Errorf("verification failed: cname_status = '%s'", respBody.Results.CNAMEStatus)
	}

	return nil
}

// CheckTrackingDomainCertificateEligibility reports whether SparkPost will
// issue a managed TLS certificate for the domain, without changing anything.
func (c *SparkPostClient) CheckTrackingDomainCertificateEligibility(domain string, subaccount int) (bool, error) {
	endpoint := fmt.Sprintf("tracking-domains/%s/certificate/check", domain)

	req, err := c.newRequest("POST", endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("failed to build certificate eligibility request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return false, fmt.Errorf("certificate eligibility check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var respBody struct {
		Results struct {
			SupportsManagedCertificate bool `json:"supportsManagedCertificate"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return false, fmt.Errorf("failed to parse certificate eligibility response: %w", err)
	}

	return respBody.Results.SupportsManagedCertificate, nil
}

// EnableTrackingDomainManagedCertificate initiates Let's Encrypt issuance for
// the tracking domain. SparkPost has no endpoint to disable it again.
func (c *SparkPostClient) EnableTrackingDomainManagedCertificate(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("tracking-domains/%s/certificate/enable", domain)

	req, err := c.newRequest("POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to build enable managed certificate request: %w", err)
	}

	if subaccount > 0 {
		req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return fmt.Errorf("enable managed certificate request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

var ErrTrackingDomainNotFound = fmt.Errorf("tracking domain not found")
