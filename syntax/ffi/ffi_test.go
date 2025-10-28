package main

import (
	"encoding/json"
	"testing"
)

func TestBuildFileEnvelopeSuccess(t *testing.T) {
	src := []byte("// comment\nfoo = \"bar\"\n")
	envelope := buildFileEnvelope("test.alloy", src)

	if envelope.ResultKind != "file" {
		t.Fatalf("unexpected result kind: %s", envelope.ResultKind)
	}
	if envelope.Status != statusOK {
		t.Fatalf("expected statusOK, got %v", envelope.Status)
	}
	if len(envelope.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(envelope.Diagnostics))
	}
	if envelope.File == nil {
		t.Fatalf("expected file payload")
	}
	if got := len(envelope.File.Body); got != 1 {
		t.Fatalf("expected 1 statement, got %d", got)
	}
	stmt := envelope.File.Body[0]
	if stmt.Kind != "attribute" {
		t.Fatalf("expected attribute statement, got %s", stmt.Kind)
	}
	if stmt.Attribute == nil || stmt.Attribute.Name.Name != "foo" {
		t.Fatalf("unexpected attribute payload: %#v", stmt.Attribute)
	}
	if stmt.Attribute.Value == nil || stmt.Attribute.Value.Kind != "literal" {
		t.Fatalf("unexpected value payload: %#v", stmt.Attribute.Value)
	}
	if len(envelope.File.Comments) != 1 {
		t.Fatalf("expected 1 comment group, got %d", len(envelope.File.Comments))
	}
}

func TestBuildExpressionEnvelopeError(t *testing.T) {
	envelope := buildExpressionEnvelope("foo(")
	if envelope.ResultKind != "expression" {
		t.Fatalf("unexpected result kind: %s", envelope.ResultKind)
	}
	if envelope.Status != statusErrors {
		t.Fatalf("expected statusErrors, got %v", envelope.Status)
	}
	if len(envelope.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics")
	}
	if envelope.Expression != nil {
		t.Fatalf("expected nil expression when parse fails")
	}
}

func TestEnvelopeJSONEncoding(t *testing.T) {
	envelope := buildExpressionEnvelope("1 + 2")
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("failed to marshal envelope: %v", err)
	}
	var decoded parseEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if decoded.SchemaVersion != schemaVersion {
		t.Fatalf("unexpected schema version: %d", decoded.SchemaVersion)
	}
}
