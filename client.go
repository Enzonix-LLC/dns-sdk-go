package enzonix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://api.ns.enzonix.com"
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "enzonix-dns-sdk-go/0.1.0"
)

// Option allows customising a Client instance produced by NewClient.
type Option func(*Client) error

// Client is an HTTP client for the Enzonix DNS API.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new Enzonix DNS API client.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("enzonix: api key must not be empty")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("enzonix: parse default base url: %w", err)
	}

	client := &Client{
		baseURL:   baseURL,
		apiKey:    apiKey,
		userAgent: defaultUserAgent,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	if client.httpClient == nil {
		client.httpClient = &http.Client{Timeout: defaultTimeout}
	}

	return client, nil
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(rawURL string) Option {
	return func(c *Client) error {
		if strings.TrimSpace(rawURL) == "" {
			return errors.New("enzonix: base url must not be empty")
		}
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("enzonix: parse base url: %w", err)
		}
		if !parsed.IsAbs() {
			return errors.New("enzonix: base url must be absolute")
		}
		c.baseURL = parsed
		return nil
	}
}

// WithHTTPClient sets a custom http.Client. It must not be nil.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return errors.New("enzonix: http client must not be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithUserAgent overrides the default user agent string.
func WithUserAgent(ua string) Option {
	return func(c *Client) error {
		c.userAgent = strings.TrimSpace(ua)
		return nil
	}
}

// APIError represents an error returned by the Enzonix API.
type APIError struct {
	StatusCode int             `json:"-"`
	Message    string          `json:"message,omitempty"`
	Code       string          `json:"code,omitempty"`
	Raw        json.RawMessage `json:"raw,omitempty"`
}

// Error satisfies the error interface.
func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	if e.Code != "" {
		return fmt.Sprintf("enzonix: %s (status=%d, code=%s)", msg, e.StatusCode, e.Code)
	}
	return fmt.Sprintf("enzonix: %s (status=%d)", msg, e.StatusCode)
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body any) (*http.Request, error) {
	if ctx == nil {
		return nil, errors.New("enzonix: context must not be nil")
	}

	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("enzonix: parse path: %w", err)
	}

	u := c.baseURL.ResolveReference(rel)
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, fmt.Errorf("enzonix: encode request body: %w", err)
		}
		bodyReader = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("enzonix: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("enzonix: request failed: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("enzonix: read response: %w", err)
	}

	if res.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: res.StatusCode}
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, apiErr); err != nil {
				// be tolerant to plain string errors
				apiErr.Message = strings.TrimSpace(string(bodyBytes))
			} else {
				apiErr.Raw = bodyBytes
			}
		}
		return apiErr
	}

	if out == nil || len(bodyBytes) == 0 {
		return nil
	}

	if err := json.Unmarshal(bodyBytes, out); err != nil {
		return fmt.Errorf("enzonix: decode response: %w", err)
	}

	return nil
}
