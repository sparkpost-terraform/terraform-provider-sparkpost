package main

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

func (c *SparkPostClient) GetTrackingDomain(domain string) (*TrackingDomain, error) {
	endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
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

func (c *SparkPostClient) DeleteTrackingDomain(domain string) error {
	endpoint := fmt.Sprintf("tracking-domains/%s", domain)

	req, err := c.newRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
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

var TrackingDomainNotFound = fmt.Errorf("tracking domain not found")
