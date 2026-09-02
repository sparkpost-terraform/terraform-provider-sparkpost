package provider

import (
	"encoding/json"
	"fmt"
)

type Subaccount struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	IPPool string `json:"ip_pool"`
}

func (c *SparkPostClient) ListSubaccounts() ([]Subaccount, error) {
	req, err := c.newRequest("GET", "subaccounts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return nil, fmt.Errorf("subaccounts request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Results []Subaccount `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode subaccounts: %w", err)
	}

	return body.Results, nil
}

// CreateSubaccount creates a subaccount and returns its assigned ID. The
// create response only carries the ID (and API key details we don't request
// via setup_api_key), never the full subaccount object, so callers must
// follow up with GetSubaccount to read back name/status/ip_pool.
func (c *SparkPostClient) CreateSubaccount(name string, ipPool string) (int, error) {
	body := map[string]interface{}{
		"name":          name,
		"setup_api_key": false,
	}
	if ipPool != "" {
		body["ip_pool"] = ipPool
	}

	req, err := c.newRequest("POST", "subaccounts", body)
	if err != nil {
		return 0, fmt.Errorf("failed to build create subaccount request: %w", err)
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		return 0, fmt.Errorf("create subaccount request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var respBody struct {
		Results struct {
			SubaccountID int `json:"subaccount_id"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return 0, fmt.Errorf("failed to parse create subaccount response: %w", err)
	}

	return respBody.Results.SubaccountID, nil
}

func (c *SparkPostClient) GetSubaccount(id int) (*Subaccount, error) {
	endpoint := fmt.Sprintf("subaccounts/%d", id)

	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build get subaccount request: %w", err)
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrSubaccountNotFound
		}
		return nil, fmt.Errorf("get subaccount request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Results Subaccount `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse get subaccount response: %w", err)
	}

	return &body.Results, nil
}

// UpdateSubaccount sets the subaccount's name, status and IP pool. SparkPost
// has no delete endpoint for subaccounts; callers terminate one by passing
// status "terminated", which is permanent.
func (c *SparkPostClient) UpdateSubaccount(id int, name string, status string, ipPool string) error {
	endpoint := fmt.Sprintf("subaccounts/%d", id)

	body := map[string]interface{}{
		"name":    name,
		"status":  status,
		"ip_pool": ipPool,
	}

	req, err := c.newRequest("PUT", endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to build update subaccount request: %w", err)
	}

	resp, err := c.doRequest(req, 200)
	if err != nil {
		if isNotFound(err) {
			return ErrSubaccountNotFound
		}
		return fmt.Errorf("update subaccount request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

var ErrSubaccountNotFound = fmt.Errorf("subaccount not found")
