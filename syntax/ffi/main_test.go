package main

import (
	"encoding/json"
	"testing"
)

func TestParseFileToJSON(t *testing.T) {
	src := []byte("// comment\nfoo = 1\n")

	payload, status := parseFileToJSON("test.alloy", src)
	if status != statusOK {
		t.Fatalf("expected statusOK, got %v", status)
	}

	var result parseResultEnvelope
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if result.SchemaVersion != schemaVersion {
		t.Fatalf("expected schema version %d, got %d", schemaVersion, result.SchemaVersion)
	}

	if result.File == nil {
		t.Fatalf("expected file in result")
	}

	if result.File.Name != "test.alloy" {
		t.Fatalf("expected filename test.alloy, got %q", result.File.Name)
	}

	if len(result.File.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result.File.Body))
	}

	stmt := result.File.Body[0]
	if stmt.Kind != "Attribute" {
		t.Fatalf("expected attribute statement, got %q", stmt.Kind)
	}

	if stmt.Attribute == nil || stmt.Attribute.Name == nil {
		t.Fatalf("expected attribute name")
	}

	if stmt.Attribute.Name.Name != "foo" {
		t.Fatalf("expected attribute name foo, got %q", stmt.Attribute.Name.Name)
	}

	if stmt.Attribute.Value == nil || stmt.Attribute.Value.Kind != "Literal" {
		t.Fatalf("expected literal value, got %#v", stmt.Attribute.Value)
	}

	if len(result.File.Comments) == 0 {
		t.Fatalf("expected serialized comments")
	}
}

func TestParseFileToJSONErrors(t *testing.T) {
	payload, status := parseFileToJSON("broken.alloy", []byte("foo =\n"))
	if status != statusError {
		t.Fatalf("expected statusError, got %v", status)
	}

	var result parseResultEnvelope
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(result.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics for broken file")
	}

	sawError := false
	for _, d := range result.Diagnostics {
		if d.Severity == "error" {
			sawError = true
		}
	}

	if !sawError {
		t.Fatalf("expected error severity in diagnostics: %#v", result.Diagnostics)
	}
}

func TestParseExpressionToJSON(t *testing.T) {
	payload, status := parseExpressionToJSON([]byte("1 + 2"))
	if status != statusOK {
		t.Fatalf("expected statusOK, got %v", status)
	}

	var result parseResultEnvelope
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if result.Expression == nil {
		t.Fatalf("expected expression in payload")
	}

	if result.Expression.Kind != "Binary" {
		t.Fatalf("expected binary expression, got %q", result.Expression.Kind)
	}
}
