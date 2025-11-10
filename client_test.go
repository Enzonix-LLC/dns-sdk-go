package enzonix

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	t.Parallel()

	client, err := NewClient("apikey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.apiKey != "apikey" {
		t.Fatalf("expected api key to be stored")
	}

	if client.baseURL.String() != defaultBaseURL {
		t.Fatalf("expected default base url, got %s", client.baseURL)
	}

	if client.httpClient == nil || client.httpClient.Timeout != defaultTimeout {
		t.Fatalf("expected default http client with timeout %v", defaultTimeout)
	}

	if client.userAgent != defaultUserAgent {
		t.Fatalf("unexpected user agent: %s", client.userAgent)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	t.Parallel()

	customURL := "https://example.test/custom"
	customUA := "custom-agent"
	customHTTP := &http.Client{Timeout: 2 * time.Second}

	client, err := NewClient("key",
		WithBaseURL(customURL),
		WithUserAgent(customUA),
		WithHTTPClient(customHTTP),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.baseURL.String() != customURL {
		t.Fatalf("expected base url %s, got %s", customURL, client.baseURL.String())
	}
	if client.userAgent != customUA {
		t.Fatalf("expected user agent %q, got %q", customUA, client.userAgent)
	}
	if client.httpClient != customHTTP {
		t.Fatalf("expected custom http client")
	}
}

func TestNewClientValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewClient(""); err == nil {
		t.Fatalf("expected error for empty api key")
	}

	if _, err := NewClient("key", WithBaseURL("://bad")); err == nil {
		t.Fatalf("expected error for invalid base url")
	}

	if _, err := NewClient("key", WithHTTPClient(nil)); err == nil {
		t.Fatalf("expected error for nil http client")
	}
}

func TestNewRequest(t *testing.T) {
	t.Parallel()

	client, err := NewClient("key", WithBaseURL("https://example.test/api"))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	query := url.Values{"q": []string{"dns"}}
	req, err := client.newRequest(context.Background(), http.MethodGet, "/zones/example.com/records", query, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := req.Header.Get("Authorization"); got != "Bearer key" {
		t.Fatalf("missing authorization header")
	}
	if got := req.Header.Get("User-Agent"); got != client.userAgent {
		t.Fatalf("missing user agent header")
	}
	if req.URL.String() != "https://example.test/zones/example.com/records?q=dns" {
		t.Fatalf("unexpected url %s", req.URL.String())
	}
}

func TestDoHandlesHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"oops","code":"bad_request"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	req, err := client.newRequest(context.Background(), http.MethodGet, "/fail", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = client.do(req, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected api error, got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "bad_request" {
		t.Fatalf("expected code bad_request, got %s", apiErr.Code)
	}
	if apiErr.Message != "oops" {
		t.Fatalf("expected message oops, got %s", apiErr.Message)
	}
}
