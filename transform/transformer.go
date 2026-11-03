package transform

import (
	"fmt"

	"github.com/palash-kulkarni/record-to-record-sync-service/model"
)

// SchemaTransformer maps an InternalRecord to a provider-specific ExternalRecord.
// One implementation per CRM provider. Stateless: same input produces same output.
type SchemaTransformer interface {
	Transform(record model.InternalRecord) (model.ExternalRecord, error)
	ProviderID() string
}

// SalesforceTransformer maps internal records to Salesforce's object model.
//
// Maps: first_name to FirstName, last_name to LastName, email to Email, phone to Phone
type SalesforceTransformer struct{}

func (t *SalesforceTransformer) ProviderID() string { return "salesforce" }

func (t *SalesforceTransformer) Transform(record model.InternalRecord) (model.ExternalRecord, error) {
	objectType, err := t.mapObjectType(record.RecordType)
	if err != nil {
		return model.ExternalRecord{}, err
	}

	fields, err := t.mapFields(record.RecordType, record.Fields)
	if err != nil {
		return model.ExternalRecord{}, err
	}

	return model.ExternalRecord{
		ExternalID: record.ID,
		ObjectType: objectType,
		Fields:     fields,
	}, nil
}

func (t *SalesforceTransformer) mapObjectType(recordType string) (string, error) {
	switch recordType {
	case "contact":
		return "Contact", nil
	case "lead":
		return "Lead", nil
	case "opportunity":
		return "Opportunity", nil
	case "account":
		return "Account", nil
	default:
		return "", fmt.Errorf("unsupported recordType for Salesforce: %s", recordType)
	}
}

func (t *SalesforceTransformer) mapFields(recordType string, internal map[string]interface{}) (map[string]interface{}, error) {
	sf := make(map[string]interface{})

	switch recordType {
	case "contact", "lead":
		if _, ok := internal["email"]; !ok {
			return nil, fmt.Errorf("email is required for Salesforce %s", recordType)
		}
		setIfPresent(sf, "FirstName", internal, "first_name")
		setIfPresent(sf, "LastName", internal, "last_name")
		setIfPresent(sf, "Email", internal, "email")
		setIfPresent(sf, "Phone", internal, "phone")

	case "opportunity":
		setIfPresent(sf, "Name", internal, "name")
		setIfPresent(sf, "Amount", internal, "amount")
		setIfPresent(sf, "StageName", internal, "stage")
		setIfPresent(sf, "CloseDate", internal, "close_date")

	case "account":
		setIfPresent(sf, "Name", internal, "company_name")
		setIfPresent(sf, "Industry", internal, "industry")
		setIfPresent(sf, "Website", internal, "website")
	}

	return sf, nil
}

// HubSpotTransformer maps internal records to HubSpot's object model.
//
// Maps: first_name to firstname, last_name to lastname, email to email, phone to phone
type HubSpotTransformer struct{}

func (t *HubSpotTransformer) ProviderID() string { return "hubspot" }

func (t *HubSpotTransformer) Transform(record model.InternalRecord) (model.ExternalRecord, error) {
	objectType, err := t.mapObjectType(record.RecordType)
	if err != nil {
		return model.ExternalRecord{}, err
	}

	fields := t.mapFields(record.RecordType, record.Fields)

	return model.ExternalRecord{
		ExternalID: record.ID,
		ObjectType: objectType,
		Fields:     fields,
	}, nil
}

func (t *HubSpotTransformer) mapObjectType(recordType string) (string, error) {
	switch recordType {
	case "contact":
		return "contacts", nil
	case "account", "company":
		return "companies", nil
	case "opportunity", "deal":
		return "deals", nil
	default:
		return "", fmt.Errorf("unsupported recordType for HubSpot: %s", recordType)
	}
}

func (t *HubSpotTransformer) mapFields(recordType string, internal map[string]interface{}) map[string]interface{} {
	hs := make(map[string]interface{})

	switch recordType {
	case "contact":
		setIfPresent(hs, "firstname", internal, "first_name")
		setIfPresent(hs, "lastname", internal, "last_name")
		setIfPresent(hs, "email", internal, "email")
		setIfPresent(hs, "phone", internal, "phone")

	case "account", "company":
		setIfPresent(hs, "name", internal, "company_name")
		setIfPresent(hs, "industry", internal, "industry")
		setIfPresent(hs, "website", internal, "website")

	case "opportunity", "deal":
		setIfPresent(hs, "dealname", internal, "name")
		setIfPresent(hs, "amount", internal, "amount")
		setIfPresent(hs, "dealstage", internal, "stage")
		setIfPresent(hs, "closedate", internal, "close_date")
	}

	return hs
}

func setIfPresent(dst map[string]interface{}, dstKey string, src map[string]interface{}, srcKey string) {
	if v, ok := src[srcKey]; ok && v != nil {
		dst[dstKey] = v
	}
}
