package provider_test

import (
	"testing"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
	"github.com/palash-kulkarni/record-to-record-sync-service/provider"
)

var contactRecord = model.ExternalRecord{
	ExternalID: "ext-1",
	ObjectType: "Contact",
	Fields:     map[string]interface{}{"Email": "a@b.com", "FirstName": "Alice"},
}

// SalesforceCRMProvider

func TestSalesforce_Create_Success(t *testing.T) {
	p := provider.NewSalesforceProvider(false)
	result := p.Create("corr-1", contactRecord)
	assertSuccess(t, result)
}

func TestSalesforce_Update_Success(t *testing.T) {
	p := provider.NewSalesforceProvider(false)
	assertSuccess(t, p.Update("corr-1", contactRecord))
}

func TestSalesforce_Delete_Success(t *testing.T) {
	p := provider.NewSalesforceProvider(false)
	assertSuccess(t, p.Delete("corr-1", contactRecord))
}

func TestSalesforce_Read_Success(t *testing.T) {
	p := provider.NewSalesforceProvider(false)
	assertSuccess(t, p.Read("corr-1", contactRecord))
}

func TestSalesforce_Create_SimulatedFailure(t *testing.T) {
	p := provider.NewSalesforceProvider(true)
	result := p.Create("corr-fail", contactRecord)
	if result.Status != model.StatusProviderError {
		t.Errorf("expected PROVIDER_ERROR, got %s", result.Status)
	}
	if result.CorrelationID != "corr-fail" {
		t.Errorf("expected correlationID corr-fail, got %s", result.CorrelationID)
	}
}

func TestSalesforce_Update_SimulatedFailure(t *testing.T) {
	p := provider.NewSalesforceProvider(true)
	result := p.Update("corr-fail", contactRecord)
	if result.Status != model.StatusProviderError {
		t.Errorf("expected PROVIDER_ERROR, got %s", result.Status)
	}
}

func TestSalesforce_ProviderID(t *testing.T) {
	p := provider.NewSalesforceProvider(false)
	if p.ProviderID() != "salesforce" {
		t.Errorf("expected salesforce, got %s", p.ProviderID())
	}
}

// HubSpotCRMProvider

var hubspotRecord = model.ExternalRecord{
	ExternalID: "ext-2",
	ObjectType: "contacts",
	Fields:     map[string]interface{}{"email": "b@b.com", "firstname": "Bob"},
}

func TestHubSpot_Create_Success(t *testing.T) {
	p := provider.NewHubSpotProvider(false)
	assertSuccess(t, p.Create("corr-2", hubspotRecord))
}

func TestHubSpot_Update_Success(t *testing.T) {
	p := provider.NewHubSpotProvider(false)
	assertSuccess(t, p.Update("corr-2", hubspotRecord))
}

func TestHubSpot_Delete_Success(t *testing.T) {
	p := provider.NewHubSpotProvider(false)
	assertSuccess(t, p.Delete("corr-2", hubspotRecord))
}

func TestHubSpot_Read_Success(t *testing.T) {
	p := provider.NewHubSpotProvider(false)
	assertSuccess(t, p.Read("corr-2", hubspotRecord))
}

func TestHubSpot_Create_SimulatedFailure(t *testing.T) {
	p := provider.NewHubSpotProvider(true)
	result := p.Create("corr-fail", hubspotRecord)
	if result.Status != model.StatusProviderError {
		t.Errorf("expected PROVIDER_ERROR, got %s", result.Status)
	}
}

func TestHubSpot_ProviderID(t *testing.T) {
	p := provider.NewHubSpotProvider(false)
	if p.ProviderID() != "hubspot" {
		t.Errorf("expected hubspot, got %s", p.ProviderID())
	}
}

func assertSuccess(t *testing.T, result model.SyncResult) {
	t.Helper()
	if result.Status != model.StatusSuccess {
		t.Errorf("expected SUCCESS, got %s: %s", result.Status, result.Message)
	}
}
