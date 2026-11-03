package transform_test

import (
	"testing"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
	"github.com/palash-kulkarni/record-to-record-sync-service/transform"
)

func TestSalesforceTransformer_Contact(t *testing.T) {
	tr := &transform.SalesforceTransformer{}
	record := model.InternalRecord{
		ID: "r1", AccountID: "a1", RecordType: "contact",
		Fields: map[string]interface{}{
			"first_name": "Alice",
			"last_name":  "Smith",
			"email":      "alice@example.com",
			"phone":      "+1-555-0100",
		},
	}

	ext, err := tr.Transform(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext.ObjectType != "Contact" {
		t.Errorf("expected Contact, got %s", ext.ObjectType)
	}
	if ext.Fields["Email"] != "alice@example.com" {
		t.Errorf("expected Email=alice@example.com, got %v", ext.Fields["Email"])
	}
	if ext.Fields["FirstName"] != "Alice" {
		t.Errorf("expected FirstName=Alice, got %v", ext.Fields["FirstName"])
	}
}

func TestSalesforceTransformer_MissingEmail(t *testing.T) {
	tr := &transform.SalesforceTransformer{}
	record := model.InternalRecord{
		ID: "r1", AccountID: "a1", RecordType: "contact",
		Fields: map[string]interface{}{"first_name": "Alice"},
	}

	_, err := tr.Transform(record)
	if err == nil {
		t.Fatal("expected error for missing email, got nil")
	}
}

func TestSalesforceTransformer_UnsupportedType(t *testing.T) {
	tr := &transform.SalesforceTransformer{}
	record := model.InternalRecord{
		ID: "r1", AccountID: "a1", RecordType: "invoice",
		Fields: map[string]interface{}{"amount": 100},
	}

	_, err := tr.Transform(record)
	if err == nil {
		t.Fatal("expected error for unsupported record type")
	}
}

func TestHubSpotTransformer_Contact(t *testing.T) {
	tr := &transform.HubSpotTransformer{}
	record := model.InternalRecord{
		ID: "r2", AccountID: "a1", RecordType: "contact",
		Fields: map[string]interface{}{
			"first_name": "Bob",
			"last_name":  "Jones",
			"email":      "bob@example.com",
		},
	}

	ext, err := tr.Transform(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext.ObjectType != "contacts" {
		t.Errorf("expected contacts, got %s", ext.ObjectType)
	}
	if ext.Fields["email"] != "bob@example.com" {
		t.Errorf("expected email=bob@example.com, got %v", ext.Fields["email"])
	}
	// HubSpot uses lowercase, should not have Salesforce-style PascalCase
	if _, ok := ext.Fields["Email"]; ok {
		t.Error("HubSpot fields should not contain PascalCase 'Email'")
	}
}

func TestHubSpotTransformer_UnsupportedType(t *testing.T) {
	tr := &transform.HubSpotTransformer{}
	record := model.InternalRecord{
		ID: "r1", AccountID: "a1", RecordType: "invoice",
		Fields: map[string]interface{}{"amount": 100},
	}

	_, err := tr.Transform(record)
	if err == nil {
		t.Fatal("expected error for unsupported record type")
	}
}
