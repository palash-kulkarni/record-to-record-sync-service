# Record-to-Record Sync Service

A Go implementation of the outbound sync pipeline for a bi-directional record synchronisation service.

See [DESIGN.md](DESIGN.md) for the full problem framing, assumptions, and design decisions.

## Prerequisites

- Go 1.23 or later (`go version`)

No external dependencies. The module uses only the Go standard library.

## Run the demo

```bash
go run cmd/main.go
```

Expected output (abbreviated):

```
Salesforce: CREATE contact
  status=SUCCESS correlationID=corr-001

HubSpot: UPDATE contact
  status=SUCCESS correlationID=corr-002

Salesforce: DELETE opportunity
  status=SUCCESS correlationID=corr-003

Validation error: missing email for Salesforce contact
  status=TRANSFORM_ERROR correlationID=corr-004 message=transform failed: email is required for Salesforce contact

Unknown provider
  status=TRANSFORM_ERROR correlationID=corr-005 message=no transformer registered for provider: zoho

Input validation error: missing record ID
  status=VALIDATION_ERROR correlationID=corr-006 message=input validation failed: record.ID is required
```

## Run all tests

```bash
go test ./...
```

Expected output:

```
ok  github.com/example/sync/pipeline
ok  github.com/example/sync/provider
ok  github.com/example/sync/ratelimit
ok  github.com/example/sync/transform
ok  github.com/example/sync/validation
```

## Run a specific package's tests

```bash
go test ./ratelimit/...
go test ./transform/...
go test ./validation/...
go test ./provider/...
go test ./pipeline/...
```

## Package layout

```
record-to-record-sync-service/
  model/        SyncEvent, InternalRecord, ExternalRecord, SyncResult, SyncAction
  validation/   input (event) and output (external record) validators
  transform/    SchemaTransformer interface; SalesforceTransformer, HubSpotTransformer
  ratelimit/    token-bucket rate limiter per (providerID, accountID)
  provider/     CRMProvider interface; mock Salesforce and HubSpot providers
  pipeline/     SyncPipeline orchestrates all steps
  cmd/          main.go demo
  DESIGN.md     design document, assumptions, trade-offs
  README.md     this file
```
