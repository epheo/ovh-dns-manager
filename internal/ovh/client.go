package ovh

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ovh-dns-manager/internal/config"
)

const (
	EndpointOVHEU = "https://eu.api.ovh.com/1.0"
	EndpointOVHCA = "https://ca.api.ovh.com/1.0"
	EndpointOVHUS = "https://api.ovhcloud.com/1.0"
)

type Client struct {
	endpoint          string
	applicationKey    string
	applicationSecret string
	consumerKey       string
	httpClient        *http.Client
}

func NewClient(creds *config.OVHCredentials) (*Client, error) {
	endpoint := EndpointOVHEU
	switch creds.Endpoint {
	case "ovh-eu":
		endpoint = EndpointOVHEU
	case "ovh-ca":
		endpoint = EndpointOVHCA
	case "ovh-us":
		endpoint = EndpointOVHUS
	default:
		if strings.HasPrefix(creds.Endpoint, "http") {
			endpoint = creds.Endpoint
		}
	}

	timeout := time.Duration(creds.Timeout) * time.Second
	if creds.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		endpoint:          endpoint,
		applicationKey:    creds.ApplicationKey,
		applicationSecret: creds.ApplicationSecret,
		consumerKey:       creds.ConsumerKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// generateSignature creates the OVH API signature using SHA1
// Note: SHA1 usage is required by the OVH API specification and cannot be changed
func (c *Client) generateSignature(method, url, body string, timestamp int64) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%s+%s+%s+%s+%s+%d",
		c.applicationSecret,
		c.consumerKey,
		method,
		url,
		body,
		timestamp,
	)))
	return fmt.Sprintf("$1$%x", h.Sum(nil))
}

func (c *Client) prepareRequest(method, path, body string) (*http.Request, error) {
	url := c.endpoint + path
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	signature := c.generateSignature(method, url, body, timestamp)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ovh-Application", c.applicationKey)
	req.Header.Set("X-Ovh-Consumer", c.consumerKey)
	req.Header.Set("X-Ovh-Signature", signature)
	req.Header.Set("X-Ovh-Timestamp", strconv.FormatInt(timestamp, 10))

	return req, nil
}

func (c *Client) doRequest(method, path, body string) (*http.Response, error) {
	req, err := c.prepareRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to read the response body for more detailed error information
		defer resp.Body.Close()
		if respBody, readErr := io.ReadAll(resp.Body); readErr == nil && len(respBody) > 0 {
			return resp, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
		}
		return resp, fmt.Errorf("API error: %s", resp.Status)
	}

	return resp, nil
}