package pipeline_test

import (
	"testing"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
	"github.com/palash-kulkarni/record-to-record-sync-service/pipeline"
	"github.com/palash-kulkarni/record-to-record-sync-service/provider"
	"github.com/palash-kulkarni/record-to-record-sync-service/ratelimit"
	"github.com/palash-kulkarni/record-to-record-sync-service/transform"
	"github.com/palash-kulkarni/record-to-record-sync-service/validation"
)

func buildPipeline(sfFailure, hsFailure bool) *pipeline.SyncPipeline {
	transformers := []transform.SchemaTransformer{
		&transform.SalesforceTransformer{},
		&transform.HubSpotTransformer{},
	}
	providers := []provider.CRMProvider{
		provider.NewSalesforceProvider(sfFailure),
		provider.NewHubSpotProvider(hsFailure),
	}
	outputValidator := validation.NewExternalRecordValidator(map[string][]string{
		"Contact":  {"Email", "FirstName"},
		"contacts": {"email", "firstname"},
	})
	rateLimiter := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{
		"salesforce": {RequestsPerSecond: 100, BurstCapacity: 150},
		"hubspot":    {RequestsPerSecond: 50, BurstCapacity: 80},
	})
	return pipeline.NewSyncPipeline(transformers, providers, outputValidator, rateLimiter)
}

func validContactEvent(providerID, correlationID string) model.SyncEvent {
	return model.SyncEvent{
		CorrelationID: correlationID,
		ProviderID:    providerID,
		Action:        model.ActionCreate,
		Record: model.InternalRecord{
			ID: "rec-1", AccountID: "acct-1", RecordType: "contact",
			Fields: map[string]interface{}{
				"first_name": "Alice",
				"last_name":  "Smith",
				"email":      "alice@example.com",
			},
		},
	}
}

func TestProcess_SalesforceCreateSuccess(t *testing.T) {
	pipe := buildPipeline(false, false)
	result := pipe.Process(validContactEvent("salesforce", "corr-1"))

	if result.Status != model.StatusSuccess {
		t.Errorf("expected SUCCESS, got %s: %s", result.Status, result.Message)
	}
	if result.CorrelationID != "corr-1" {
		t.Errorf("expected correlationID corr-1, got %s", result.CorrelationID)
	}
}

func TestProcess_HubSpotCreateSuccess(t *testing.T) {
	pipe := buildPipeline(false, false)
	result := pipe.Process(validContactEvent("hubspot", "corr-2"))

	if result.Status != model.StatusSuccess {
		t.Errorf("expected SUCCESS, got %s: %s", result.Status, result.Message)
	}
}

func TestProcess_InputValidation_MissingRecordID(t *testing.T) {
	pipe := buildPipeline(false, false)
	event := validContactEvent("salesforce", "corr-3")
	event.Record.ID = ""

	result := pipe.Process(event)
	if result.Status != model.StatusValidationError {
		t.Errorf("expected VALIDATION_ERROR, got %s", result.Status)
	}
}

func TestProcess_InputValidation_MissingProviderID(t *testing.T) {
	pipe := buildPipeline(false, false)
	event := validContactEvent("", "corr-4")

	result := pipe.Process(event)
	if result.Status != model.StatusValidationError {
		t.Errorf("expected VALIDATION_ERROR, got %s", result.Status)
	}
}

func TestProcess_TransformError_MissingEmail(t *testing.T) {
	pipe := buildPipeline(false, false)
	event := model.SyncEvent{
		CorrelationID: "corr-5",
		ProviderID:    "salesforce",
		Action:        model.ActionCreate,
		Record: model.InternalRecord{
			ID: "rec-5", AccountID: "acct-1", RecordType: "contact",
			Fields: map[string]interface{}{
				"first_name": "No Email",
			},
		},
	}

	result := pipe.Process(event)
	if result.Status != model.StatusTransformError {
		t.Errorf("expected TRANSFORM_ERROR, got %s: %s", result.Status, result.Message)
	}
}

func TestProcess_UnknownProvider(t *testing.T) {
	pipe := buildPipeline(false, false)
	event := validContactEvent("zoho", "corr-6")

	result := pipe.Process(event)
	if result.Status != model.StatusTransformError {
		t.Errorf("expected TRANSFORM_ERROR for unknown provider, got %s", result.Status)
	}
}

func TestProcess_ProviderError(t *testing.T) {
	pipe := buildPipeline(true, false) // Salesforce simulates failure
	result := pipe.Process(validContactEvent("salesforce", "corr-7"))

	if result.Status != model.StatusProviderError {
		t.Errorf("expected PROVIDER_ERROR, got %s", result.Status)
	}
}

func TestProcess_RateLimited(t *testing.T) {
	// Very tight rate limit to force exhaustion
	rateLimiter := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{
		"salesforce": {RequestsPerSecond: 1, BurstCapacity: 1},
	})
	pipe := pipeline.NewSyncPipeline(
		[]transform.SchemaTransformer{&transform.SalesforceTransformer{}},
		[]provider.CRMProvider{provider.NewSalesforceProvider(false)},
		validation.NewExternalRecordValidator(map[string][]string{
			"Contact": {"Email", "FirstName"},
		}),
		rateLimiter,
	)

	// First call should succeed (uses the 1 burst token)
	r1 := pipe.Process(validContactEvent("salesforce", "corr-8a"))
	if r1.Status != model.StatusSuccess {
		t.Errorf("first call expected SUCCESS, got %s", r1.Status)
	}

	// Second call immediately after should be rate-limited
	r2 := pipe.Process(validContactEvent("salesforce", "corr-8b"))
	if r2.Status != model.StatusRateLimited {
		t.Errorf("second call expected RATE_LIMITED, got %s", r2.Status)
	}
}

func TestProcess_AllActions(t *testing.T) {
	pipe := buildPipeline(false, false)
	actions := []model.SyncAction{
		model.ActionCreate, model.ActionUpdate,
		model.ActionDelete, model.ActionRead,
	}

	for _, action := range actions {
		event := validContactEvent("salesforce", "corr-action")
		event.Action = action
		result := pipe.Process(event)
		if result.Status != model.StatusSuccess {
			t.Errorf("action %s: expected SUCCESS, got %s: %s", action, result.Status, result.Message)
		}
	}
}
