package validation

import (
	"fmt"
	"strings"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
)

// ValidationResult carries the outcome of a validation check.
type ValidationResult struct {
	Valid  bool
	Errors []string
}

func ok() ValidationResult {
	return ValidationResult{Valid: true}
}

func invalid(errors []string) ValidationResult {
	return ValidationResult{Valid: false, Errors: errors}
}

func (v ValidationResult) String() string {
	if v.Valid {
		return "OK"
	}
	return "INVALID: " + strings.Join(v.Errors, "; ")
}

// ValidateEvent is the input gate.
// Catches malformed events before any transformation happens.
func ValidateEvent(event model.SyncEvent) ValidationResult {
	var errors []string

	if event.CorrelationID == "" {
		errors = append(errors, "correlationID is required")
	}
	if event.ProviderID == "" {
		errors = append(errors, "providerID is required")
	}
	if event.Action == "" {
		errors = append(errors, "action is required")
	}
	if event.Record.ID == "" {
		errors = append(errors, "record.ID is required")
	}
	if event.Record.AccountID == "" {
		errors = append(errors, "record.AccountID is required")
	}
	if event.Record.RecordType == "" {
		errors = append(errors, "record.RecordType is required")
	}
	if len(event.Record.Fields) == 0 {
		errors = append(errors, "record.Fields must not be empty")
	}

	if len(errors) > 0 {
		return invalid(errors)
	}
	return ok()
}

// ExternalRecordValidator is the output gate.
// Validates the transformed payload against the provider's schema contract
// before dispatching to the external API.
type ExternalRecordValidator struct {
	// requiredFieldsByObjectType maps objectType to a list of required field names.
	requiredFieldsByObjectType map[string][]string
}

func NewExternalRecordValidator(requiredFields map[string][]string) *ExternalRecordValidator {
	return &ExternalRecordValidator{requiredFieldsByObjectType: requiredFields}
}

func (v *ExternalRecordValidator) Validate(record model.ExternalRecord) ValidationResult {
	var errors []string

	if record.ObjectType == "" {
		errors = append(errors, "objectType is required")
	}
	if record.Fields == nil {
		errors = append(errors, "fields must not be nil")
	}

	if record.ObjectType != "" && record.Fields != nil {
		required := v.requiredFieldsByObjectType[record.ObjectType]
		for _, field := range required {
			if _, exists := record.Fields[field]; !exists {
				errors = append(errors, fmt.Sprintf("required field missing: %s", field))
			}
		}
	}

	if len(errors) > 0 {
		return invalid(errors)
	}
	return ok()
}
