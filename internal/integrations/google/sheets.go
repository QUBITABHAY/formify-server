package google

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"
)

type FormField struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

func ParseFormSchema(schemaJSON []byte) ([]FormField, error) {
	log.Printf("Parsing form schema: %s", string(schemaJSON))

	var schema []FormField
	if err := json.Unmarshal(schemaJSON, &schema); err == nil && len(schema) > 0 {
		normalized := normalizeFields(schema)
		log.Printf("Parsed %d fields from direct array", len(normalized))
		for i, f := range normalized {
			log.Printf("  Field %d: ID=%s, Label=%s, Name=%s", i, f.ID, f.Label, f.Name)
		}
		return normalized, nil
	}

	var schemaWithFields struct {
		Fields []FormField `json:"fields"`
	}
	if err := json.Unmarshal(schemaJSON, &schemaWithFields); err == nil && len(schemaWithFields.Fields) > 0 {
		normalized := normalizeFields(schemaWithFields.Fields)
		log.Printf("Parsed %d fields from 'fields' key", len(normalized))
		return normalized, nil
	}

	var rawSchema []map[string]interface{}
	if err := json.Unmarshal(schemaJSON, &rawSchema); err == nil {
		log.Printf("Parsing %d raw objects", len(rawSchema))
		var fields []FormField
		for _, field := range rawSchema {
			log.Printf("  Raw field: %+v", field)
			f := FormField{}
			if id, ok := field["id"].(string); ok {
				f.ID = id
			} else if id, ok := field["name"].(string); ok {
				f.ID = id
			} else if id, ok := field["fieldId"].(string); ok {
				f.ID = id
			} else if id, ok := field["key"].(string); ok {
				f.ID = id
			}

			if label, ok := field["label"].(string); ok {
				f.Label = label
			} else if label, ok := field["title"].(string); ok {
				f.Label = label
			} else if label, ok := field["name"].(string); ok && f.Label == "" {
				f.Label = label
			} else if label, ok := field["text"].(string); ok {
				f.Label = label
			} else if label, ok := field["placeholder"].(string); ok && f.Label == "" {
				f.Label = label
			}
			if typ, ok := field["type"].(string); ok {
				f.Type = typ
			}
			if f.ID != "" || f.Label != "" {
				fields = append(fields, f)
			}
		}
		normalized := normalizeFields(fields)
		log.Printf("Extracted %d fields from raw parsing", len(normalized))
		return normalized, nil
	}

	log.Printf("Failed to parse form schema")
	return nil, fmt.Errorf("failed to parse form schema")
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

	headers := []string{"Submission ID", "Submitted At"}
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
