package pipeline

import (
	"fmt"
	"strings"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
	"github.com/palash-kulkarni/record-to-record-sync-service/provider"
	"github.com/palash-kulkarni/record-to-record-sync-service/ratelimit"
	"github.com/palash-kulkarni/record-to-record-sync-service/transform"
	"github.com/palash-kulkarni/record-to-record-sync-service/validation"
)

// SyncPipeline orchestrates the outbound sync flow:
// ValidateInput, Transform, ValidateOutput, RateLimit, Dispatch
//
// It depends only on interfaces; no concrete providers or transformers are imported here.
type SyncPipeline struct {
	transformers    map[string]transform.SchemaTransformer // providerID to transformer
	providers       map[string]provider.CRMProvider        // providerID to CRM client
	outputValidator *validation.ExternalRecordValidator
	rateLimiter     *ratelimit.RateLimiter
}

// NewSyncPipeline wires the pipeline with the given components.
// Multiple transformers and providers can be registered; they are looked up
// by providerID at runtime.
func NewSyncPipeline(
	transformers []transform.SchemaTransformer,
	providers []provider.CRMProvider,
	outputValidator *validation.ExternalRecordValidator,
	rateLimiter *ratelimit.RateLimiter,
) *SyncPipeline {
	tMap := make(map[string]transform.SchemaTransformer, len(transformers))
	for _, t := range transformers {
		tMap[t.ProviderID()] = t
	}

	pMap := make(map[string]provider.CRMProvider, len(providers))
	for _, p := range providers {
		pMap[p.ProviderID()] = p
	}

	return &SyncPipeline{
		transformers:    tMap,
		providers:       pMap,
		outputValidator: outputValidator,
		rateLimiter:     rateLimiter,
	}
}

// Process runs a SyncEvent through the full outbound pipeline.
// Expected failures (rate limit, bad data) are returned as SyncResult statuses, not errors.
// Only programming errors (unknown action, nil maps) return a non-success result via the default branch.
func (p *SyncPipeline) Process(event model.SyncEvent) model.SyncResult {
	inputResult := validation.ValidateEvent(event)
	if !inputResult.Valid {
		return model.FailureResult(
			model.StatusValidationError,
			event.CorrelationID,
			"input validation failed: "+strings.Join(inputResult.Errors, "; "),
		)
	}

	transformer, ok := p.transformers[event.ProviderID]
	if !ok {
		return model.FailureResult(
			model.StatusTransformError,
			event.CorrelationID,
			fmt.Sprintf("no transformer registered for provider: %s", event.ProviderID),
		)
	}

	crmProvider, ok := p.providers[event.ProviderID]
	if !ok {
		return model.FailureResult(
			model.StatusProviderError,
			event.CorrelationID,
			fmt.Sprintf("no provider registered for: %s", event.ProviderID),
		)
	}

	externalRecord, err := transformer.Transform(event.Record)
	if err != nil {
		return model.FailureResult(
			model.StatusTransformError,
			event.CorrelationID,
			"transform failed: "+err.Error(),
		)
	}

	// Validate after transform, not before: we're checking the provider's schema contract,
	// not the internal one. Running this before transform would validate the wrong shape.
	outputResult := p.outputValidator.Validate(externalRecord)
	if !outputResult.Valid {
		return model.FailureResult(
			model.StatusValidationError,
			event.CorrelationID,
			"output validation failed: "+strings.Join(outputResult.Errors, "; "),
		)
	}

	if !p.rateLimiter.TryAcquire(event.ProviderID, event.Record.AccountID) {
		return model.FailureResult(
			model.StatusRateLimited,
			event.CorrelationID,
			fmt.Sprintf("rate limit exceeded for provider=%s account=%s", event.ProviderID, event.Record.AccountID),
		)
	}

	return p.dispatch(crmProvider, event.CorrelationID, event.Action, externalRecord)
}

func (p *SyncPipeline) dispatch(
	crm provider.CRMProvider,
	correlationID string,
	action model.SyncAction,
	record model.ExternalRecord,
) model.SyncResult {
	switch action {
	case model.ActionCreate:
		return crm.Create(correlationID, record)
	case model.ActionUpdate:
		return crm.Update(correlationID, record)
	case model.ActionDelete:
		return crm.Delete(correlationID, record)
	case model.ActionRead:
		return crm.Read(correlationID, record)
	default:
		return model.FailureResult(
			model.StatusValidationError,
			correlationID,
			fmt.Sprintf("unknown action: %s", action),
		)
	}
}
