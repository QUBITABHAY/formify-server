package google

import (
	"testing"
	"time"
)

func TestParseFormSchema_DirectArray(t *testing.T) {
	input := []byte(`[{"id":"q1","label":"Full Name","type":"text"},{"name":"email","title":"Email Address"}]`)
	fields, err := ParseFormSchema(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].ID != "q1" || fields[0].Label != "Full Name" {
		t.Errorf("field 0 mismatch: %+v", fields[0])
	}
	if fields[1].ID != "email" || fields[1].Label != "Email Address" {
		t.Errorf("field 1 normalization mismatch: %+v", fields[1])
	}
}

func TestParseFormSchema_FieldsObject(t *testing.T) {
	input := []byte(`{"fields":[{"id":"age","label":"Your Age"}]}`)
	fields, err := ParseFormSchema(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 1 || fields[0].ID != "age" {
		t.Fatalf("expected 1 field 'age', got %+v", fields)
	}
}

func TestParseFormSchema_RawObjectArray(t *testing.T) {
	input := []byte(`[{"id":12345,"fieldId":"feedback","label":"Tell us more"}]`)
	fields, err := ParseFormSchema(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 1 || fields[0].ID != "feedback" || fields[0].Label != "Tell us more" {
		t.Fatalf("expected extracted field, got %+v", fields)
	}
}

func TestParseFormSchema_InvalidJSON(t *testing.T) {
	_, err := ParseFormSchema([]byte("not a json"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseFormSchema_EmptyArray(t *testing.T) {
	_, err := ParseFormSchema([]byte("[]"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractHeaders(t *testing.T) {
	fields := []FormField{
		{ID: "q1", Label: "Name"},
		{ID: "q2", Label: ""},
	}

	headers := ExtractHeaders(fields)
	expected := []string{"Submission ID", "Submitted At", "Name", "q2"}
	if len(headers) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i := range expected {
		if headers[i] != expected[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expected[i], headers[i])
		}
	}
}

func TestResponseToRow(t *testing.T) {
	fields := []FormField{
		{ID: "name", Label: "Name"},
		{ID: "age", Label: "Age"},
		{ID: "missing", Label: "Missing Field"},
	}

	submittedAt := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	data := []byte(`{"name":"Alice","age":30}`)

	row, err := ResponseToRow(123, submittedAt, data, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(row) != 5 {
		t.Fatalf("expected 5 items, got %d", len(row))
	}
	if row[0] != "123" {
		t.Errorf("expected row[0] == '123', got %v", row[0])
	}
	if row[1] != submittedAt.Format(time.RFC3339) {
		t.Errorf("expected submittedAt RFC3339, got %v", row[1])
	}
	if row[2] != "Alice" {
		t.Errorf("expected Alice, got %v", row[2])
	}
	if row[3] != "30" {
		t.Errorf("expected 30, got %v", row[3])
	}
	if row[4] != "" {
		t.Errorf("expected empty string for missing field, got %v", row[4])
	}

	_, err = ResponseToRow(123, submittedAt, []byte("bad json"), fields)
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
}

func TestResponseToRowWithoutSchema(t *testing.T) {
	submittedAt := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	data := []byte(`{"beta":"b","alpha":"a"}`)

	row, headers, err := ResponseToRowWithoutSchema(456, submittedAt, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedHeaders := []string{"Submission ID", "Submitted At", "alpha", "beta"}
	if len(headers) != len(expectedHeaders) {
		t.Fatalf("expected %d headers, got %d", len(expectedHeaders), len(headers))
	}
	for i := range expectedHeaders {
		if headers[i] != expectedHeaders[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expectedHeaders[i], headers[i])
		}
	}

	if row[0] != "456" || row[2] != "a" || row[3] != "b" {
		t.Fatalf("unexpected row contents: %v", row)
	}

	_, _, err = ResponseToRowWithoutSchema(456, submittedAt, []byte("bad json"))
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
}

func TestSanitizeFormula(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"=SUM(A1:B1)", "'=SUM(A1:B1)"},
		{"+12345", "'+12345"},
		{"-formula", "'-formula"},
		{"@mention", "'@mention"},
		{"\tleading tab", "'\tleading tab"},
		{"\rleading cr", "'\rleading cr"},
		{"normal text", "normal text"},
		{"", ""},
	}

	for _, tt := range tests {
		got := sanitizeFormula(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeFormula(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil", nil, ""},
		{"bool true", true, "Yes"},
		{"bool false", false, "No"},
		{"float64", float64(42.5), "42.5"},
		{"string normal", "hello", "hello"},
		{"string formula", "=1+1", "'=1+1"},
		{"slice", []any{"apple", "banana"}, "apple, banana"},
		{"slice with formula", []any{"=cmd", "safe"}, "'=cmd, safe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.input)
			if got != tt.expected {
				t.Errorf("formatValue() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNormalizeFields(t *testing.T) {
	fields := []FormField{
		{Name: "field_one", Title: "Field One"},
	}
	normalized := normalizeFields(fields)
	if normalized[0].ID != "field_one" {
		t.Fatalf("expected ID to be normalized to Name, got %q", normalized[0].ID)
	}
	if normalized[0].Label != "Field One" {
		t.Fatalf("expected Label to be normalized to Title, got %q", normalized[0].Label)
	}
}
