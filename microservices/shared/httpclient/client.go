package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ServiceClient represents an HTTP client for service-to-service communication
type ServiceClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// NewServiceClient creates a new service client
func NewServiceClient(baseURL string, timeout time.Duration) *ServiceClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Request represents an HTTP request
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

// Do performs an HTTP request
func (c *ServiceClient) Do(ctx context.Context, req *Request) (*Response, error) {
	url := c.baseURL + req.Path

	var body io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonBody)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       responseBody,
		Headers:    resp.Header,
	}, nil
}

// Get performs a GET request
func (c *ServiceClient) Get(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  "GET",
		Path:    path,
		Headers: headers,
	})
}

// Post performs a POST request
func (c *ServiceClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  "POST",
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Put performs a PUT request
func (c *ServiceClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  "PUT",
		Path:    path,
		Body:    body,
		Headers: headers,
	})
}

// Delete performs a DELETE request
func (c *ServiceClient) Delete(ctx context.Context, path string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  "DELETE",
		Path:    path,
		Headers: headers,
	})
}

// UnmarshalResponse unmarshals the response body into a struct
func (r *Response) UnmarshalResponse(v interface{}) error {
	if r.StatusCode >= 400 {
		return fmt.Errorf("HTTP error %d: %s", r.StatusCode, string(r.Body))
	}

	if len(r.Body) == 0 {
		return nil
	}

	return json.Unmarshal(r.Body, v)
}

// IsSuccess returns true if the response status code indicates success
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// BaseURL returns the base URL of the service client
func (c *ServiceClient) BaseURL() string {
	return c.baseURL
}
