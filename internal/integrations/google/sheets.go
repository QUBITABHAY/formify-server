package google

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type FormField struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

func ParseFormSchema(schemaJSON []byte) ([]FormField, error) {
	var schema []FormField
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		var rawSchema []map[string]interface{}
		if err := json.Unmarshal(schemaJSON, &rawSchema); err != nil {
			return nil, fmt.Errorf("failed to parse form schema: %w", err)
		}

		for _, field := range rawSchema {
			f := FormField{}
			if id, ok := field["id"].(string); ok {
				f.ID = id
			}
			if label, ok := field["label"].(string); ok {
				f.Label = label
			}
			if typ, ok := field["type"].(string); ok {
				f.Type = typ
			}
			schema = append(schema, f)
		}
	}
	return schema, nil
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
		if value, ok := data[field.ID]; ok {
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
