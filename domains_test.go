package enzonix

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDomains(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/domains" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Domain{{ID: "domain-1", Name: "example.com."}})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	domains, err := client.ListDomains(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 1 || domains[0].ID != "domain-1" {
		t.Fatalf("unexpected domains: %#v", domains)
	}
}

func TestCreateDomain(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST")
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload["name"] != "example.com" {
			t.Fatalf("unexpected name %s", payload["name"])
		}
		json.NewEncoder(w).Encode(Domain{ID: "domain-1", Name: "example.com."})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	domain, err := client.CreateDomain(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if domain.ID != "domain-1" {
		t.Fatalf("unexpected domain: %#v", domain)
	}
}

func TestDeleteDomain(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/domains/domain-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	if err := client.DeleteDomain(context.Background(), "domain-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckNameserver(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/domains/domain-1/check-nameserver" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(NameserverCheckResponse{
			Domain: Domain{ID: "domain-1", Name: "example.com."},
			Check: struct {
				Valid  bool   `json:"valid"`
				Status string `json:"status"`
			}{Valid: true, Status: "valid"},
		})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	resp, err := client.CheckNameserver(context.Background(), "domain-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Check.Valid {
		t.Fatalf("expected valid check")
	}
}

func TestExportBindZone(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/domains/domain-1/export/bind" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "$ORIGIN example.com.\n")
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	data, err := client.ExportBindZone(context.Background(), "domain-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) == "" {
		t.Fatalf("expected data")
	}
}

func TestImportBindZone(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/import/bind" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(BindImportResponse{
			Domain:         Domain{ID: "domain-1", Name: "example.com."},
			RecordsCreated: 1,
		})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	resp, err := client.ImportBindZone(context.Background(), []byte("$ORIGIN example.com."), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RecordsCreated != 1 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestRotateAPIKey(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/rotate-api-key" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ClientProfile{ID: "client-1", APIToken: "new"})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	profile, err := client.RotateAPIKey(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.APIToken != "new" {
		t.Fatalf("unexpected token %s", profile.APIToken)
	}
}
