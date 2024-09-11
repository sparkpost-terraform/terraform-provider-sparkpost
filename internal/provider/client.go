package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SparkPostClient struct {
	APIUrl string
	APIKey string
	client *http.Client
}

func NewSparkPostClient(apiUrl string, apiKey string) *SparkPostClient {
	return &SparkPostClient{
		APIUrl: apiUrl,
		APIKey: apiKey,
		client: &http.Client{},
	}
}

func (c *SparkPostClient) newRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	var bodyBytes []byte
	var err error

	url := c.APIUrl + endpoint

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authoization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *SparkPostClient) doRequest(req *http.Request, expectedCode int) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedCode {
		defer resp.Body.Close()
		return resp, fmt.Errorf("Request failed with status: %s", resp.Status)
	}

	return resp, nil
}
