package main

import (
	"fmt"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
	"github.com/palash-kulkarni/record-to-record-sync-service/pipeline"
	"github.com/palash-kulkarni/record-to-record-sync-service/provider"
	"github.com/palash-kulkarni/record-to-record-sync-service/ratelimit"
	"github.com/palash-kulkarni/record-to-record-sync-service/transform"
	"github.com/palash-kulkarni/record-to-record-sync-service/validation"
)

func main() {
	transformers := []transform.SchemaTransformer{
		&transform.SalesforceTransformer{},
		&transform.HubSpotTransformer{},
	}

	providers := []provider.CRMProvider{
		provider.NewSalesforceProvider(false),
		provider.NewHubSpotProvider(false),
	}

	outputValidator := validation.NewExternalRecordValidator(map[string][]string{
		"Contact":     {"Email", "FirstName"},
		"contacts":    {"email", "firstname"},
		"Lead":        {"Email"},
		"Opportunity": {"Name", "StageName"},
	})

	rateLimiter := ratelimit.NewRateLimiter(map[string]ratelimit.ProviderRateConfig{
		"salesforce": {RequestsPerSecond: 100, BurstCapacity: 150},
		"hubspot":    {RequestsPerSecond: 50, BurstCapacity: 80},
	})

	pipe := pipeline.NewSyncPipeline(transformers, providers, outputValidator, rateLimiter)

	scenarios := []struct {
		name  string
		event model.SyncEvent
	}{
		{
			name: "Salesforce: CREATE contact",
			event: model.SyncEvent{
				CorrelationID: "corr-001",
				ProviderID:    "salesforce",
				Action:        model.ActionCreate,
				Record: model.InternalRecord{
					ID: "rec-001", AccountID: "acct-A", RecordType: "contact",
					Fields: map[string]interface{}{
						"first_name": "Alice",
						"last_name":  "Smith",
						"email":      "alice@example.com",
						"phone":      "+1-555-0100",
					},
				},
			},
		},
		{
			name: "HubSpot: UPDATE contact",
			event: model.SyncEvent{
				CorrelationID: "corr-002",
				ProviderID:    "hubspot",
				Action:        model.ActionUpdate,
				Record: model.InternalRecord{
					ID: "rec-002", AccountID: "acct-B", RecordType: "contact",
					Fields: map[string]interface{}{
						"first_name": "Bob",
						"last_name":  "Jones",
						"email":      "bob@example.com",
					},
				},
			},
		},
		{
			name: "Salesforce: DELETE opportunity",
			event: model.SyncEvent{
				CorrelationID: "corr-003",
				ProviderID:    "salesforce",
				Action:        model.ActionDelete,
				Record: model.InternalRecord{
					ID: "rec-003", AccountID: "acct-A", RecordType: "opportunity",
					Fields: map[string]interface{}{
						"name": "Big Deal", "amount": 50000, "stage": "Negotiation", "close_date": "2026-12-31",
					},
				},
			},
		},
		{
			name: "Validation error: missing email for Salesforce contact",
			event: model.SyncEvent{
				CorrelationID: "corr-004",
				ProviderID:    "salesforce",
				Action:        model.ActionCreate,
				Record: model.InternalRecord{
					ID: "rec-004", AccountID: "acct-A", RecordType: "contact",
					Fields: map[string]interface{}{
						"first_name": "No Email",
					},
				},
			},
		},
		{
			name: "Unknown provider",
			event: model.SyncEvent{
				CorrelationID: "corr-005",
				ProviderID:    "zoho",
				Action:        model.ActionCreate,
				Record: model.InternalRecord{
					ID: "rec-005", AccountID: "acct-C", RecordType: "contact",
					Fields: map[string]interface{}{"email": "x@x.com"},
				},
			},
		},
		{
			name: "Input validation error: missing record ID",
			event: model.SyncEvent{
				CorrelationID: "corr-006",
				ProviderID:    "salesforce",
				Action:        model.ActionCreate,
				Record: model.InternalRecord{
					ID: "", AccountID: "acct-A", RecordType: "contact",
					Fields: map[string]interface{}{"email": "x@x.com"},
				},
			},
		},
	}

	for _, s := range scenarios {
		fmt.Printf("\n%s\n", s.name)
		result := pipe.Process(s.event)
		fmt.Printf("  status=%s correlationID=%s", result.Status, result.CorrelationID)
		if result.Message != "" {
			fmt.Printf(" message=%s", result.Message)
		}
		fmt.Println()
	}
}
