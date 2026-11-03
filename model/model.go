package model

// SyncAction is the CRUD operation to perform on the external system.
type SyncAction string

const (
	ActionCreate SyncAction = "CREATE"
	ActionUpdate SyncAction = "UPDATE"
	ActionDelete SyncAction = "DELETE"
	ActionRead   SyncAction = "READ"
)

// SyncStatus is the outcome of a pipeline execution.
type SyncStatus string

const (
	StatusSuccess         SyncStatus = "SUCCESS"
	StatusTransformError  SyncStatus = "TRANSFORM_ERROR"  // bad data, dead-letter, do not retry
	StatusValidationError SyncStatus = "VALIDATION_ERROR" // schema violation, dead-letter
	StatusRateLimited     SyncStatus = "RATE_LIMITED"     // back-pressure, retry with backoff
	StatusProviderError   SyncStatus = "PROVIDER_ERROR"   // external API failure, retry
)

// InternalRecord is the canonical internal representation.
// Field names follow our internal schema, which differs from any external CRM.
type InternalRecord struct {
	ID         string
	AccountID  string
	RecordType string // "contact", "lead", "opportunity", "account"
	Fields     map[string]interface{}
}

// ExternalRecord is the provider-specific representation after transformation.
// Field names and structure match what the external CRM API expects.
type ExternalRecord struct {
	ExternalID string // ID in the external system; empty on CREATE
	ObjectType string // CRM object type e.g. "Contact", "contacts"
	Fields     map[string]interface{}
}

// SyncEvent is a single synchronisation request from the internal event stream.
// One event = one record change = one outbound API call.
type SyncEvent struct {
	CorrelationID string // unique ID for tracing and deduplication hints
	ProviderID    string // target CRM: "salesforce", "hubspot", etc.
	Action        SyncAction
	Record        InternalRecord
}

// SyncResult is the outcome of processing a SyncEvent through the pipeline.
// Expected failures (rate limit, validation) are returned as results, not errors,
// so callers can branch without unwrapping error chains.
type SyncResult struct {
	Status        SyncStatus
	CorrelationID string
	Message       string // human-readable detail; empty on SUCCESS
}

func SuccessResult(correlationID string) SyncResult {
	return SyncResult{Status: StatusSuccess, CorrelationID: correlationID}
}

func FailureResult(status SyncStatus, correlationID, message string) SyncResult {
	return SyncResult{Status: status, CorrelationID: correlationID, Message: message}
}

func (r SyncResult) IsSuccess() bool {
	return r.Status == StatusSuccess
}
