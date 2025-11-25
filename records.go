package enzonix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const clientAPIPrefix = "/api/client"

// Domain represents a domain resource managed via the client API.
type Domain struct {
	ID                    string     `json:"id"`
	ClientID              string     `json:"client_id"`
	Name                  string     `json:"name"`
	Active                bool       `json:"active"`
	CreatedAt             *time.Time `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at"`
	NameserverLastChecked *time.Time `json:"nameserver_last_checked_at"`
	NameserverVerifiedAt  *time.Time `json:"nameserver_verified_at"`
	NameserverCheckStatus string     `json:"nameserver_check_status"`
}

// NameserverCheckResponse represents the check-nameserver payload.
type NameserverCheckResponse struct {
	Domain Domain `json:"domain"`
	Check  struct {
		Valid  bool   `json:"valid"`
		Status string `json:"status"`
	} `json:"check"`
}

// Record describes a DNS record managed by Enzonix.
type Record struct {
	ID           string     `json:"id"`
	DomainID     string     `json:"domain_id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	TTL          int        `json:"ttl"`
	CountryCodes []string   `json:"country_codes"`
	Priority     int        `json:"priority"`
	Value        string     `json:"value"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

// CreateRecordRequest defines the payload used to create a new record.
type CreateRecordRequest struct {
	DomainID     string   `json:"domain_id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Value        string   `json:"value"`
	TTL          *int     `json:"ttl,omitempty"`
	Priority     *int     `json:"priority,omitempty"`
	CountryCodes []string `json:"country_codes,omitempty"`
}

// UpdateRecordRequest defines the payload used to update an existing record.
type UpdateRecordRequest struct {
	Name         *string  `json:"name,omitempty"`
	Type         *string  `json:"type,omitempty"`
	Value        *string  `json:"value,omitempty"`
	TTL          *int     `json:"ttl,omitempty"`
	Priority     *int     `json:"priority,omitempty"`
	CountryCodes []string `json:"country_codes,omitempty"`
}

// ListDomains retrieves all domains owned by the authenticated client.
func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	req, err := c.newRequest(ctx, http.MethodGet, clientAPIPrefix+"/domains", nil, nil)
	if err != nil {
		return nil, err
	}

	var domains []Domain
	if err := c.do(req, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

// CreateDomain creates a new domain for the authenticated client.
func (c *Client) CreateDomain(ctx context.Context, name string) (*Domain, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("enzonix: domain name must not be empty")
	}

	payload := map[string]string{"name": name}
	req, err := c.newRequest(ctx, http.MethodPost, clientAPIPrefix+"/domains", nil, payload)
	if err != nil {
		return nil, err
	}

	var domain Domain
	if err := c.do(req, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// DeleteDomain deletes a domain by ID.
func (c *Client) DeleteDomain(ctx context.Context, domainID string) error {
	if err := requireID(domainID, "domain id"); err != nil {
		return err
	}

	path := fmt.Sprintf("%s/domains/%s", clientAPIPrefix, url.PathEscape(domainID))
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// CheckNameserver triggers a nameserver validation for a domain.
func (c *Client) CheckNameserver(ctx context.Context, domainID string) (*NameserverCheckResponse, error) {
	if err := requireID(domainID, "domain id"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/domains/%s/check-nameserver", clientAPIPrefix, url.PathEscape(domainID))
	req, err := c.newRequest(ctx, http.MethodPost, path, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp NameserverCheckResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListDomainRecords returns all records for the given domain ID.
func (c *Client) ListDomainRecords(ctx context.Context, domainID string) ([]Record, error) {
	if err := requireID(domainID, "domain id"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/domains/%s/records", clientAPIPrefix, url.PathEscape(domainID))
	req, err := c.newRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}

	var records []Record
	if err := c.do(req, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// CreateRecord creates a record and returns the created resource.
func (c *Client) CreateRecord(ctx context.Context, payload CreateRecordRequest) (*Record, error) {
	if err := requireID(payload.DomainID, "domain id"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Name) == "" {
		return nil, fmt.Errorf("enzonix: record name must not be empty")
	}
	if strings.TrimSpace(payload.Type) == "" {
		return nil, fmt.Errorf("enzonix: record type must not be empty")
	}
	if strings.TrimSpace(payload.Value) == "" {
		return nil, fmt.Errorf("enzonix: record value must not be empty")
	}

	req, err := c.newRequest(ctx, http.MethodPost, clientAPIPrefix+"/records", nil, payload)
	if err != nil {
		return nil, err
	}

	var record Record
	if err := c.do(req, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// UpdateRecord updates a record by ID and returns the updated resource.
func (c *Client) UpdateRecord(ctx context.Context, recordID string, payload UpdateRecordRequest) (*Record, error) {
	if err := requireID(recordID, "record id"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/records/%s", clientAPIPrefix, url.PathEscape(recordID))
	req, err := c.newRequest(ctx, http.MethodPut, path, nil, payload)
	if err != nil {
		return nil, err
	}

	var record Record
	if err := c.do(req, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// DeleteRecord deletes a record by ID.
func (c *Client) DeleteRecord(ctx context.Context, recordID string) error {
	if err := requireID(recordID, "record id"); err != nil {
		return err
	}

	path := fmt.Sprintf("%s/records/%s", clientAPIPrefix, url.PathEscape(recordID))
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// ExportBindZone downloads a domain's records as a BIND zone file.
func (c *Client) ExportBindZone(ctx context.Context, domainID string) ([]byte, error) {
	if err := requireID(domainID, "domain id"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/domains/%s/export/bind", clientAPIPrefix, url.PathEscape(domainID))
	req, err := c.newRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/plain")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("enzonix: request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, parseAPIError(res)
	}

	return io.ReadAll(io.LimitReader(res.Body, 4<<20))
}

// BindImportResponse represents the response from the BIND import endpoint.
type BindImportResponse struct {
	Domain         Domain   `json:"domain"`
	RecordsCreated int      `json:"records_created"`
	Records        []Record `json:"records"`
	PartialSuccess bool     `json:"partial_success"`
	Errors         []string `json:"errors"`
}

// ImportBindZone imports records from a BIND zone file.
func (c *Client) ImportBindZone(ctx context.Context, zoneData []byte, contentType string) (*BindImportResponse, error) {
	if len(zoneData) == 0 {
		return nil, fmt.Errorf("enzonix: zone data must not be empty")
	}
	if contentType == "" {
		contentType = "text/plain"
	}

	req, err := c.newRequest(ctx, http.MethodPost, clientAPIPrefix+"/import/bind", nil, nil)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(zoneData))
	req.ContentLength = int64(len(zoneData))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(zoneData)), nil
	}
	req.Header.Set("Content-Type", contentType)

	var resp BindImportResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// RotateAPIKey rotates the client's API key and returns the updated profile.
type ClientProfile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	APIToken    string    `json:"api_token"`
	DomainLimit int       `json:"domain_limit"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *Client) RotateAPIKey(ctx context.Context) (*ClientProfile, error) {
	req, err := c.newRequest(ctx, http.MethodPost, clientAPIPrefix+"/rotate-api-key", nil, nil)
	if err != nil {
		return nil, err
	}

	var profile ClientProfile
	if err := c.do(req, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func requireID(value, label string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("enzonix: %s must not be empty", label)
	}
	return nil
}

func parseAPIError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	apiErr := &APIError{StatusCode: res.StatusCode}
	if len(body) > 0 {
		if err := json.Unmarshal(body, apiErr); err != nil {
			apiErr.Message = strings.TrimSpace(string(body))
		} else {
			apiErr.Raw = body
		}
	}
	return apiErr
}
