# Enzonix DNS SDK for Go

`github.com/Enzonix-LLC/dns-sdk-go` is a Go client for the Enzonix DNS API. It provides convenient helpers to authenticate with the API and manage DNS records programmatically.

> **Status:** experimental – the public Enzonix DNS API is still evolving. Interface changes may occur before v1.0.0.

## Installation

```bash
go get github.com/Enzonix-LLC/dns-sdk-go
```

## Quick start

```go
package main

import (
	"context"
	"log"

	"github.com/Enzonix-LLC/dns-sdk-go"
)

func main() {
	client, err := enzonix.NewClient("ENZONIX_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	// Fetch all domains available to the API token
	domains, err := client.ListDomains(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if len(domains) == 0 {
		log.Println("no domains yet")
		return
	}

	records, err := client.ListDomainRecords(context.Background(), domains[0].ID)
	if err != nil {
		log.Fatal(err)
	}
	for _, record := range records {
		log.Printf("%s %s %s", record.Name, record.Type, record.Value)
	}
}
```

### Creating records

```go
record, err := client.CreateRecord(ctx, enzonix.CreateRecordRequest{
	DomainID: domains[0].ID,
	Name:     "_acme-challenge",
	Type:     "TXT",
	Value:    "token",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("created record %s", record.ID)
```

### Updating records

```go
value := "203.0.113.42"
updated, err := client.UpdateRecord(ctx, record.ID, enzonix.UpdateRecordRequest{
	Value: &value,
})
if err != nil {
	log.Fatal(err)
}
log.Printf("updated record to %s", updated.Value)
```

### Deleting records

```go
if err := client.DeleteRecord(ctx, record.ID); err != nil {
	log.Fatal(err)
}
```

## Configuration

The client accepts functional options:

- `WithBaseURL` – point the SDK at a custom API endpoint (useful for testing or regional deployments).
- `WithHTTPClient` – provide your own `*http.Client` (for example, to set custom transport settings).
- `WithUserAgent` – override the default user-agent string.

## Development

```bash
go test ./...
```

## License

MIT © Enzonix


