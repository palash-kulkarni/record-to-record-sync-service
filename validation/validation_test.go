package validation_test

import (
	"testing"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
	"github.com/palash-kulkarni/record-to-record-sync-service/validation"
)

// ValidateEvent

func TestValidateEvent_Valid(t *testing.T) {
	event := model.SyncEvent{
		CorrelationID: "corr-1",
		ProviderID:    "salesforce",
		Action:        model.ActionCreate,
		Record: model.InternalRecord{
			ID:         "rec-1",
			AccountID:  "acct-1",
			RecordType: "contact",
			Fields:     map[string]interface{}{"email": "a@b.com"},
		},
	}
	result := validation.ValidateEvent(event)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateEvent_MissingCorrelationID(t *testing.T) {
	event := validEvent()
	event.CorrelationID = ""
	assertInvalid(t, validation.ValidateEvent(event), "correlationID")
}

func TestValidateEvent_MissingProviderID(t *testing.T) {
	event := validEvent()
	event.ProviderID = ""
	assertInvalid(t, validation.ValidateEvent(event), "providerID")
}

func TestValidateEvent_MissingAction(t *testing.T) {
	event := validEvent()
	event.Action = ""
	assertInvalid(t, validation.ValidateEvent(event), "action")
}

func TestValidateEvent_MissingRecordID(t *testing.T) {
	event := validEvent()
	event.Record.ID = ""
	assertInvalid(t, validation.ValidateEvent(event), "record.ID")
}

func TestValidateEvent_MissingAccountID(t *testing.T) {
	event := validEvent()
	event.Record.AccountID = ""
	assertInvalid(t, validation.ValidateEvent(event), "record.AccountID")
}

func TestValidateEvent_MissingRecordType(t *testing.T) {
	event := validEvent()
	event.Record.RecordType = ""
	assertInvalid(t, validation.ValidateEvent(event), "record.RecordType")
}

func TestValidateEvent_EmptyFields(t *testing.T) {
	event := validEvent()
	event.Record.Fields = map[string]interface{}{}
	assertInvalid(t, validation.ValidateEvent(event), "record.Fields")
}

func TestValidateEvent_MultipleErrors(t *testing.T) {
	event := model.SyncEvent{}
	result := validation.ValidateEvent(event)
	if result.Valid {
		t.Fatal("expected invalid for zero-value event")
	}
	if len(result.Errors) < 2 {
		t.Errorf("expected multiple errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

// ExternalRecordValidator

func TestExternalRecordValidator_Valid(t *testing.T) {
	v := validation.NewExternalRecordValidator(map[string][]string{
		"Contact": {"Email", "FirstName"},
	})
	record := model.ExternalRecord{
		ObjectType: "Contact",
		Fields:     map[string]interface{}{"Email": "a@b.com", "FirstName": "Alice"},
	}
	result := v.Validate(record)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestExternalRecordValidator_MissingObjectType(t *testing.T) {
	v := validation.NewExternalRecordValidator(nil)
	record := model.ExternalRecord{Fields: map[string]interface{}{"Email": "a@b.com"}}
	assertInvalid(t, v.Validate(record), "objectType")
}

func TestExternalRecordValidator_NilFields(t *testing.T) {
	v := validation.NewExternalRecordValidator(nil)
	record := model.ExternalRecord{ObjectType: "Contact"}
	assertInvalid(t, v.Validate(record), "fields")
}

func TestExternalRecordValidator_MissingRequiredField(t *testing.T) {
	v := validation.NewExternalRecordValidator(map[string][]string{
		"Contact": {"Email", "FirstName"},
	})
	record := model.ExternalRecord{
		ObjectType: "Contact",
		Fields:     map[string]interface{}{"Email": "a@b.com"},
	}
	result := v.Validate(record)
	if result.Valid {
		t.Fatal("expected invalid when required field is missing")
	}
	found := false
	for _, e := range result.Errors {
		if e == "required field missing: FirstName" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about FirstName, got: %v", result.Errors)
	}
}

func TestExternalRecordValidator_UnknownObjectType_PassesThrough(t *testing.T) {
	v := validation.NewExternalRecordValidator(map[string][]string{
		"Contact": {"Email"},
	})
	record := model.ExternalRecord{
		ObjectType: "Lead",
		Fields:     map[string]interface{}{"Name": "Bob"},
	}
	result := v.Validate(record)
	if !result.Valid {
		t.Errorf("object type with no required-field rule should pass, got: %v", result.Errors)
	}
}

func validEvent() model.SyncEvent {
	return model.SyncEvent{
		CorrelationID: "corr-1",
		ProviderID:    "salesforce",
		Action:        model.ActionCreate,
		Record: model.InternalRecord{
			ID:         "rec-1",
			AccountID:  "acct-1",
			RecordType: "contact",
			Fields:     map[string]interface{}{"email": "a@b.com"},
		},
	}
}

func assertInvalid(t *testing.T, result validation.ValidationResult, context string) {
	t.Helper()
	if result.Valid {
		t.Errorf("[%s] expected invalid, got valid", context)
	}
}
