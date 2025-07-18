package provider

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type TargetDomain struct {
	Domain string `json:"domain"`
}

func (c *SparkPostClient) CreateDomain(domain string, subaccount int) error {
	body := map[string]interface{}{
		"domain": domain,
	}

	req, err := c.newRequest("POST", "sending-domains", body)
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

func (c *SparkPostClient) GetDomain(domain string, subaccount int) (*TargetDomain, error) {
	endpoint := fmt.Sprintf("sending-domains/%s", domain)

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

	var targetDomain TargetDomain
	err = json.NewDecoder(resp.Body).Decode(&targetDomain)
	if err != nil {
		return nil, err
	}

	return &targetDomain, nil
}

func (c *SparkPostClient) DeleteDomain(domain string, subaccount int) error {
	endpoint := fmt.Sprintf("sending-domains/%s", domain)

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
		return DomainNotFound
	}

	defer resp.Body.Close()

	return nil
}

var DomainNotFound = fmt.Errorf("sending domain not found")
