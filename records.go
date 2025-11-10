package enzonix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Record describes a DNS record managed by Enzonix.
type Record struct {
	ID        string     `json:"id"`
	Zone      string     `json:"zone"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Content   string     `json:"content"`
	TTL       int        `json:"ttl,omitempty"`
	Priority  uint       `json:"priority,omitempty"`
	Weight    uint       `json:"weight,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// ListRecordsOptions controls the result set returned by ListRecords.
type ListRecordsOptions struct {
	Name    string
	Type    string
	Page    int
	PerPage int
}

// CreateRecordRequest defines the payload used to create a new record.
type CreateRecordRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl,omitempty"`
	Priority uint   `json:"priority,omitempty"`
	Weight   uint   `json:"weight,omitempty"`
}

// UpdateRecordRequest defines the payload used to update an existing record.
type UpdateRecordRequest struct {
	Name     *string `json:"name,omitempty"`
	Type     *string `json:"type,omitempty"`
	Content  *string `json:"content,omitempty"`
	TTL      *int    `json:"ttl,omitempty"`
	Priority *uint   `json:"priority,omitempty"`
	Weight   *uint   `json:"weight,omitempty"`
}

// ListRecords returns all records for the given zone.
func (c *Client) ListRecords(ctx context.Context, zone string, opts *ListRecordsOptions) ([]Record, error) {
	if err := c.requireZone(zone); err != nil {
		return nil, err
	}

	query := url.Values{}
	if opts != nil {
		if opts.Name != "" {
			query.Set("name", opts.Name)
		}
		if opts.Type != "" {
			query.Set("type", opts.Type)
		}
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(opts.PerPage))
		}
	}

	req, err := c.newRequest(ctx, http.MethodGet, zoneRecordsPath(zone), query, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Records []Record `json:"records"`
	}

	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response.Records, nil
}

// CreateRecord creates a record and returns the created resource.
func (c *Client) CreateRecord(ctx context.Context, zone string, payload CreateRecordRequest) (*Record, error) {
	if err := c.requireZone(zone); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Name) == "" {
		return nil, fmt.Errorf("enzonix: record name must not be empty")
	}
	if strings.TrimSpace(payload.Type) == "" {
		return nil, fmt.Errorf("enzonix: record type must not be empty")
	}
	if strings.TrimSpace(payload.Content) == "" {
		return nil, fmt.Errorf("enzonix: record content must not be empty")
	}

	req, err := c.newRequest(ctx, http.MethodPost, zoneRecordsPath(zone), nil, payload)
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
func (c *Client) UpdateRecord(ctx context.Context, zone, recordID string, payload UpdateRecordRequest) (*Record, error) {
	if err := c.requireZone(zone); err != nil {
		return nil, err
	}
	if strings.TrimSpace(recordID) == "" {
		return nil, fmt.Errorf("enzonix: record id must not be empty")
	}

	req, err := c.newRequest(ctx, http.MethodPatch, zoneRecordPath(zone, recordID), nil, payload)
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
func (c *Client) DeleteRecord(ctx context.Context, zone, recordID string) error {
	if err := c.requireZone(zone); err != nil {
		return err
	}
	if strings.TrimSpace(recordID) == "" {
		return fmt.Errorf("enzonix: record id must not be empty")
	}

	req, err := c.newRequest(ctx, http.MethodDelete, zoneRecordPath(zone, recordID), nil, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) requireZone(zone string) error {
	if strings.TrimSpace(zone) == "" {
		return fmt.Errorf("enzonix: zone must not be empty")
	}
	return nil
}

func zoneRecordsPath(zone string) string {
	return fmt.Sprintf("/zones/%s/records", url.PathEscape(trimZone(zone)))
}

func zoneRecordPath(zone, recordID string) string {
	return fmt.Sprintf("%s/%s", zoneRecordsPath(zone), url.PathEscape(recordID))
}

func trimZone(zone string) string {
	zone = strings.TrimSpace(zone)
	return strings.TrimSuffix(zone, ".")
}
