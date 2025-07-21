package provider

import (
	"encoding/json"
	"fmt"
)

type Subaccount struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
	defer resp.Body.Close()

	var body struct {
		Results []Subaccount `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode subaccounts: %w", err)
	}

	return body.Results, nil
}