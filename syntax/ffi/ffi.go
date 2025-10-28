package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct {
    char* data;
    int64_t length;
    int32_t status;
} AlloyParseResult;
*/
import "C"

import (
	"encoding/json"
	"errors"
	"unsafe"

	"github.com/grafana/alloy/syntax/diag"
	"github.com/grafana/alloy/syntax/parser"
)

type parseStatus int32

const (
	statusOK parseStatus = iota
	statusWarnings
	statusErrors
)

const schemaVersion = 1

//export AlloyParseFileJSON
func AlloyParseFileJSON(filename *C.char, data *C.char) C.AlloyParseResult {
	goFilename := C.GoString(filename)
	goData := C.GoString(data)

	envelope := buildFileEnvelope(goFilename, []byte(goData))
	return marshalEnvelope(envelope)
}

//export AlloyParseExpressionJSON
func AlloyParseExpressionJSON(expression *C.char) C.AlloyParseResult {
	goExpr := C.GoString(expression)

	envelope := buildExpressionEnvelope(goExpr)
	return marshalEnvelope(envelope)
}

//export AlloyFree
func AlloyFree(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}

func marshalEnvelope(envelope parseEnvelope) C.AlloyParseResult {
	payload, err := json.Marshal(envelope)
	if err != nil {
		// Return an explicit error result with no payload.
		return buildErrorResult(err)
	}

	cBytes := C.CBytes(payload)
	return C.AlloyParseResult{
		data:   (*C.char)(cBytes),
		length: C.int64_t(len(payload)),
		status: C.int32_t(envelope.Status),
	}
}

func buildErrorResult(err error) C.AlloyParseResult {
	payload, _ := json.Marshal(parseEnvelope{
		SchemaVersion: schemaVersion,
		Status:        statusErrors,
		Diagnostics: []serializedDiagnostic{
			{
				Severity: "error",
				Message:  err.Error(),
			},
		},
	})

	cBytes := C.CBytes(payload)
	return C.AlloyParseResult{
		data:   (*C.char)(cBytes),
		length: C.int64_t(len(payload)),
		status: C.int32_t(statusErrors),
	}
}

type parseEnvelope struct {
	SchemaVersion int                    `json:"schemaVersion"`
	ResultKind    string                 `json:"resultKind"`
	File          *serializedFile        `json:"file,omitempty"`
	Expression    *serializedExpr        `json:"expression,omitempty"`
	Diagnostics   []serializedDiagnostic `json:"diagnostics"`
	Status        parseStatus            `json:"status"`
}

func buildFileEnvelope(filename string, data []byte) parseEnvelope {
	result := parseEnvelope{
		SchemaVersion: schemaVersion,
		ResultKind:    "file",
	}

	file, err := parser.ParseFile(filename, data)
	var diags diag.Diagnostics
	if err != nil {
		var d diag.Diagnostics
		if errors.As(err, &d) {
			diags = d
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  err.Error(),
			})
		}
	}

	if file != nil {
		serialized := serializeFile(file)
		result.File = &serialized
	}

	result.Diagnostics = serializeDiagnostics(diags)
	result.Status = computeStatus(diags)
	return result
}

func buildExpressionEnvelope(expr string) parseEnvelope {
	result := parseEnvelope{
		SchemaVersion: schemaVersion,
		ResultKind:    "expression",
	}

	parsed, err := parser.ParseExpression(expr)
	var diags diag.Diagnostics
	if err != nil {
		var d diag.Diagnostics
		if errors.As(err, &d) {
			diags = d
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  err.Error(),
			})
		}
	}

	if parsed != nil {
		serialized := serializeExpr(parsed)
		result.Expression = serialized
	}

	result.Diagnostics = serializeDiagnostics(diags)
	result.Status = computeStatus(diags)
	return result
}

func computeStatus(diags diag.Diagnostics) parseStatus {
	if len(diags) == 0 {
		return statusOK
	}
	if diags.HasErrors() {
		return statusErrors
	}
	return statusWarnings
}
