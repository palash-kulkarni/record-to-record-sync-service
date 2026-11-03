package provider

import (
	"fmt"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
)

// CRMProvider abstracts a single external CRM system.
// One implementation per provider. Pipeline depends only on this interface.
type CRMProvider interface {
	Create(correlationID string, record model.ExternalRecord) model.SyncResult
	Update(correlationID string, record model.ExternalRecord) model.SyncResult
	Delete(correlationID string, record model.ExternalRecord) model.SyncResult
	Read(correlationID string, record model.ExternalRecord) model.SyncResult
	ProviderID() string
}

// SalesforceCRMProvider is a mock Salesforce provider.
type SalesforceCRMProvider struct {
	simulateFailure bool
}

func NewSalesforceProvider(simulateFailure bool) *SalesforceCRMProvider {
	return &SalesforceCRMProvider{simulateFailure: simulateFailure}
}

func (p *SalesforceCRMProvider) ProviderID() string { return "salesforce" }

func (p *SalesforceCRMProvider) Create(correlationID string, record model.ExternalRecord) model.SyncResult {
	if p.simulateFailure {
		return model.FailureResult(model.StatusProviderError, correlationID, "Salesforce: UNABLE_TO_LOCK_ROW")
	}
	fmt.Printf("[Salesforce] CREATE %s id=%s fields=%v\n", record.ObjectType, record.ExternalID, record.Fields)
	return model.SuccessResult(correlationID)
}

func (p *SalesforceCRMProvider) Update(correlationID string, record model.ExternalRecord) model.SyncResult {
	if p.simulateFailure {
		return model.FailureResult(model.StatusProviderError, correlationID, "Salesforce: REQUEST_LIMIT_EXCEEDED")
	}
	fmt.Printf("[Salesforce] UPDATE %s id=%s fields=%v\n", record.ObjectType, record.ExternalID, record.Fields)
	return model.SuccessResult(correlationID)
}

func (p *SalesforceCRMProvider) Delete(correlationID string, record model.ExternalRecord) model.SyncResult {
	fmt.Printf("[Salesforce] DELETE %s id=%s\n", record.ObjectType, record.ExternalID)
	return model.SuccessResult(correlationID)
}

func (p *SalesforceCRMProvider) Read(correlationID string, record model.ExternalRecord) model.SyncResult {
	fmt.Printf("[Salesforce] READ %s id=%s\n", record.ObjectType, record.ExternalID)
	return model.SuccessResult(correlationID)
}

// HubSpotCRMProvider is a mock HubSpot provider.
type HubSpotCRMProvider struct {
	simulateFailure bool
}

func NewHubSpotProvider(simulateFailure bool) *HubSpotCRMProvider {
	return &HubSpotCRMProvider{simulateFailure: simulateFailure}
}

func (p *HubSpotCRMProvider) ProviderID() string { return "hubspot" }

func (p *HubSpotCRMProvider) Create(correlationID string, record model.ExternalRecord) model.SyncResult {
	if p.simulateFailure {
		return model.FailureResult(model.StatusProviderError, correlationID, "HubSpot: 429 Too Many Requests")
	}
	fmt.Printf("[HubSpot] CREATE %s id=%s fields=%v\n", record.ObjectType, record.ExternalID, record.Fields)
	return model.SuccessResult(correlationID)
}

func (p *HubSpotCRMProvider) Update(correlationID string, record model.ExternalRecord) model.SyncResult {
	fmt.Printf("[HubSpot] UPDATE %s id=%s fields=%v\n", record.ObjectType, record.ExternalID, record.Fields)
	return model.SuccessResult(correlationID)
}

func (p *HubSpotCRMProvider) Delete(correlationID string, record model.ExternalRecord) model.SyncResult {
	fmt.Printf("[HubSpot] DELETE %s id=%s\n", record.ObjectType, record.ExternalID)
	return model.SuccessResult(correlationID)
}

func (p *HubSpotCRMProvider) Read(correlationID string, record model.ExternalRecord) model.SyncResult {
	fmt.Printf("[HubSpot] READ %s id=%s\n", record.ObjectType, record.ExternalID)
	return model.SuccessResult(correlationID)
}
