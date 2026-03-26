package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"
)

const fixedResponseHeaders = 2

var errFailedToParseFormSchema = errors.New("failed to parse form schema")

type FormField struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

func ParseFormSchema(schemaJSON []byte) ([]FormField, error) {
	log.Printf("Parsing form schema: %s", string(schemaJSON))

	// Try parsing as direct array of FormField
	fields, err := parseDirectArray(schemaJSON)
	if err == nil && len(fields) > 0 {
		return normalizeFields(fields), nil
	}

	// Try parsing as object with "fields" key
	fields, err = parseFieldsObject(schemaJSON)
	if err == nil && len(fields) > 0 {
		return normalizeFields(fields), nil
	}

	// Try parsing as raw objects
	fields, err = parseRawObjects(schemaJSON)
	if err == nil && len(fields) > 0 {
		return normalizeFields(fields), nil
	}

	log.Printf("Failed to parse form schema")
	return nil, errFailedToParseFormSchema
}

func parseDirectArray(schemaJSON []byte) ([]FormField, error) {
	var schema []FormField
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, err
	}
	if len(schema) > 0 {
		log.Printf("Parsed %d fields from direct array", len(schema))
		for i, f := range schema {
			log.Printf("  Field %d: ID=%s, Label=%s, Name=%s", i, f.ID, f.Label, f.Name)
		}
	}
	return schema, nil
}

func parseFieldsObject(schemaJSON []byte) ([]FormField, error) {
	var schemaWithFields struct {
		Fields []FormField `json:"fields"`
	}
	if err := json.Unmarshal(schemaJSON, &schemaWithFields); err != nil {
		return nil, err
	}
	if len(schemaWithFields.Fields) > 0 {
		log.Printf("Parsed %d fields from 'fields' key", len(schemaWithFields.Fields))
	}
	return schemaWithFields.Fields, nil
}

func parseRawObjects(schemaJSON []byte) ([]FormField, error) {
	var rawSchema []map[string]interface{}
	if err := json.Unmarshal(schemaJSON, &rawSchema); err != nil {
		return nil, err
	}

	log.Printf("Parsing %d raw objects", len(rawSchema))
	fields := make([]FormField, 0, len(rawSchema))
	for _, field := range rawSchema {
		f := extractFieldFromRaw(field)
		if f.ID != "" || f.Label != "" {
			fields = append(fields, f)
		}
	}

	if len(fields) > 0 {
		log.Printf("Extracted %d fields from raw parsing", len(fields))
	}
	return fields, nil
}

func extractFieldFromRaw(field map[string]interface{}) FormField {
	log.Printf("  Raw field: %+v", field)
	f := FormField{}

	// Extract ID
	f.ID = extractStringField(field, []string{"id", "name", "fieldId", "key"})

	// Extract Label
	f.Label = extractLabelFromRaw(field)

	// Extract Type
	if typ, ok := field["type"].(string); ok {
		f.Type = typ
	}

	return f
}

func extractStringField(field map[string]interface{}, keys []string) string {
	for _, key := range keys {
		if val, ok := field[key].(string); ok && val != "" {
			return val
		}
	}
	return ""
}

func extractLabelFromRaw(field map[string]interface{}) string {
	return extractStringField(field, []string{"label", "title", "text", "name", "placeholder"})
}

func normalizeFields(fields []FormField) []FormField {
	for i := range fields {
		if fields[i].ID == "" && fields[i].Name != "" {
			fields[i].ID = fields[i].Name
		}

		if fields[i].Label == "" && fields[i].Title != "" {
			fields[i].Label = fields[i].Title
		}

		if fields[i].Label == "" && fields[i].Name != "" {
			fields[i].Label = fields[i].Name
		}
	}
	return fields
}

func ExtractHeaders(fields []FormField) []string {
	headers := []string{"Submission ID", "Submitted At"}
	for _, field := range fields {
		if field.Label != "" {
			headers = append(headers, field.Label)
		} else if field.ID != "" {
			headers = append(headers, field.ID)
		}
	}
	return headers
}

func ResponseToRow(responseID int32, submittedAt time.Time, responseData []byte, fields []FormField) ([]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(responseData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response data: %w", err)
	}

	row := []interface{}{
		fmt.Sprintf("%d", responseID),
		submittedAt.Format(time.RFC3339),
	}

	for _, field := range fields {
		var value interface{}
		var found bool

		if field.ID != "" {
			value, found = data[field.ID]
		}

		if !found && field.Name != "" {
			value, found = data[field.Name]
		}

		if !found && field.Label != "" {
			value, found = data[field.Label]
		}

		if found {
			row = append(row, formatValue(value))
		} else {
			row = append(row, "")
		}
	}

	return row, nil
}

func ResponseToRowWithoutSchema(responseID int32, submittedAt time.Time, responseData []byte) ([]interface{}, []string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(responseData, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response data: %w", err)
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headers := make([]string, 0, fixedResponseHeaders+len(keys))
	headers = append(headers, "Submission ID", "Submitted At")
	headers = append(headers, keys...)

	row := []interface{}{
		fmt.Sprintf("%d", responseID),
		submittedAt.Format(time.RFC3339),
	}

	for _, key := range keys {
		row = append(row, formatValue(data[key]))
	}

	return row, headers, nil
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case []interface{}:
		result := ""
		for i, item := range v {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%v", item)
		}
		return result
	default:
		return fmt.Sprintf("%v", v)
	}
}
