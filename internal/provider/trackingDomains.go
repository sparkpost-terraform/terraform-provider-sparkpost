package provider

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type TrackingDomain struct {
	Domain string `json:"domain"`
}

func (c *SparkPostClient) CreateTrackingDomain(domain string, https bool, subaccount int) error {
	body := map[string]interface{}{
		"domain": domain,
		"secure": https,
	}

	req, err := c.newRequest("POST", "tracking-domains", body)
	if err != nil {
		return err
	}

	if subaccount > 0 {
	   req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *SparkPostClient) GetTrackingDomain(domain string, subaccount int) (*TrackingDomain, error) {
	endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if subaccount > 0 {
	   req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var trackingDomain TrackingDomain
	err = json.NewDecoder(resp.Body).Decode(&trackingDomain)
	if err != nil {
		return nil, err
	}

	return &trackingDomain, nil
}

func (c *SparkPostClient) DeleteTrackingDomain(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	if subaccount > 0 {
	   req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 204)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return TrackingDomainNotFound
	}

	defer resp.Body.Close()

	return nil
}

func (c *SparkPostClient) UpdateTrackingDomain(domain string, https bool, subaccount int) error {
	body := map[string]interface{}{
		"secure": https,
	}

    endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("PUT", endpoint, body)
	if err != nil {
		return err
	}

	if subaccount > 0 {
	   req.Header.Set("X-MSYS-SUBACCOUNT", strconv.Itoa(subaccount))
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

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
	defer resp.Body.Close()

	var respBody struct {
		Results struct {
			Verified    bool   `json:"verified"`
			CNAMEStatus string `json:"cname_status"`			
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("failed to parse verification response: %w", err)
	}

	if respBody.Results.Verified != true {
		return fmt.Errorf("verification failed: cname_status = '%s'", respBody.Results.CNAMEStatus)
	}

	return nil
}

var TrackingDomainNotFound = fmt.Errorf("tracking domain not found")