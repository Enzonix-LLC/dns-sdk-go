package enzonix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDomainRecords(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET got %s", r.Method)
		}
		if r.URL.Path != "/api/client/domains/domain-123/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer key" {
			t.Fatalf("missing authorization header")
		}

		json.NewEncoder(w).Encode([]Record{
			{ID: "1", DomainID: "domain-123", Name: "www", Type: "A", Value: "1.1.1.1"},
		})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	records, err := client.ListDomainRecords(context.Background(), "domain-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 || records[0].ID != "1" {
		t.Fatalf("unexpected records: %#v", records)
	}
}

func TestCreateRecord(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST got %s", r.Method)
		}
		if r.URL.Path != "/api/client/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload CreateRecordRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload.DomainID != "domain-123" {
			t.Fatalf("unexpected domain id: %s", payload.DomainID)
		}
		json.NewEncoder(w).Encode(Record{
			ID:       "abc",
			DomainID: payload.DomainID,
			Name:     payload.Name,
			Type:     payload.Type,
			Value:    payload.Value,
		})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	record, err := client.CreateRecord(context.Background(), CreateRecordRequest{
		DomainID: "domain-123",
		Name:     "_acme-challenge",
		Type:     "TXT",
		Value:    "token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.ID != "abc" {
		t.Fatalf("unexpected record: %#v", record)
	}
}

func TestUpdateRecord(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT got %s", r.Method)
		}
		if r.URL.Path != "/api/client/records/abc" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Record{ID: "abc", DomainID: "domain-123", Name: "www", Type: "A", Value: "2.2.2.2"})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	value := "2.2.2.2"
	record, err := client.UpdateRecord(context.Background(), "abc", UpdateRecordRequest{
		Value: &value,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.Value != "2.2.2.2" {
		t.Fatalf("unexpected record: %#v", record)
	}
}

func TestDeleteRecord(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE got %s", r.Method)
		}
		if r.URL.Path != "/api/client/records/abc" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	if err := client.DeleteRecord(context.Background(), "abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
