package provider

import (
	"bytes"
	"encoding/json"
	"errors"
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

	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// apiError carries the HTTP status of a failed request so callers can
// distinguish a 404 (not found) from other failures without string-matching
// the error message.
type apiError struct {
	StatusCode int
	Status     string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("request failed with status: %s", e.Status)
}

// isNotFound reports whether err is an apiError with a 404 status.
func isNotFound(err error) bool {
	var apiErr *apiError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func (c *SparkPostClient) doRequest(req *http.Request, expectedCode int) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedCode {
		defer func() { _ = resp.Body.Close() }()
		return resp, &apiError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	return resp, nil
}
