package enzonix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRecords(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET got %s", r.Method)
		}
		if r.URL.Path != "/zones/example.com/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "www" {
			t.Fatalf("expected name query www, got %s", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer key" {
			t.Fatalf("missing authorization header")
		}

		json.NewEncoder(w).Encode(map[string]any{
			"records": []Record{
				{ID: "1", Zone: "example.com", Name: "www", Type: "A", Content: "1.1.1.1"},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	records, err := client.ListRecords(context.Background(), "example.com.", &ListRecordsOptions{Name: "www"})
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
		var payload CreateRecordRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload.Name != "_acme-challenge" {
			t.Fatalf("unexpected name: %s", payload.Name)
		}
		json.NewEncoder(w).Encode(Record{
			ID:      "abc",
			Zone:    "example.com",
			Name:    payload.Name,
			Type:    payload.Type,
			Content: payload.Content,
		})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	record, err := client.CreateRecord(context.Background(), "example.com", CreateRecordRequest{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "token",
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
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH got %s", r.Method)
		}
		if r.URL.Path != "/zones/example.com/records/abc" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Record{ID: "abc", Zone: "example.com", Name: "www", Type: "A", Content: "2.2.2.2"})
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	content := "2.2.2.2"
	record, err := client.UpdateRecord(context.Background(), "example.com.", "abc", UpdateRecordRequest{
		Content: &content,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.Content != "2.2.2.2" {
		t.Fatalf("unexpected record: %#v", record)
	}
}

func TestDeleteRecord(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient("key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	if err := client.DeleteRecord(context.Background(), "example.com", "abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
