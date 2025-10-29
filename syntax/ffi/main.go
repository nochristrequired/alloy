package main

// #include <stdlib.h>
//
// typedef struct {
// char* data;
// size_t length;
// int status;
// } AlloyJSONResult;
import "C"

import (
	"encoding/json"
	"errors"
	"unsafe"

	"github.com/grafana/alloy/syntax/ast"
	"github.com/grafana/alloy/syntax/diag"
	"github.com/grafana/alloy/syntax/parser"
	"github.com/grafana/alloy/syntax/token"
)

const (
	schemaVersion = 1
)

type resultStatus int

const (
	statusOK resultStatus = iota
	statusWarn
	statusError
	statusInternalError
)

type parseResultEnvelope struct {
	SchemaVersion int                    `json:"schemaVersion"`
	File          *serializedFile        `json:"file,omitempty"`
	Expression    *serializedExpr        `json:"expression,omitempty"`
	Diagnostics   []serializedDiagnostic `json:"diagnostics"`
}

type serializedFile struct {
	Name     string                   `json:"name"`
	Body     []serializedStmt         `json:"body"`
	Comments []serializedCommentGroup `json:"comments"`
}

type serializedCommentGroup struct {
	Comments []serializedComment `json:"comments"`
}

type serializedComment struct {
	Text  string             `json:"text"`
	Start serializedPosition `json:"start"`
}

type serializedStmt struct {
	Kind      string                   `json:"kind"`
	Attribute *serializedAttributeStmt `json:"attribute,omitempty"`
	Block     *serializedBlockStmt     `json:"block,omitempty"`
}

type serializedAttributeStmt struct {
	Name  *serializedIdent `json:"name"`
	Value *serializedExpr  `json:"value"`
}

type serializedBlockStmt struct {
	Name      []string           `json:"name"`
	NamePos   serializedPosition `json:"namePos"`
	Label     string             `json:"label,omitempty"`
	LabelPos  serializedPosition `json:"labelPos"`
	Body      []serializedStmt   `json:"body"`
	LCurlyPos serializedPosition `json:"lCurlyPos"`
	RCurlyPos serializedPosition `json:"rCurlyPos"`
}

type serializedExpr struct {
	Kind       string                    `json:"kind"`
	Secret     bool                      `json:"secret"`
	Identifier *serializedIdentifierExpr `json:"identifier,omitempty"`
	Literal    *serializedLiteralExpr    `json:"literal,omitempty"`
	Array      *serializedArrayExpr      `json:"array,omitempty"`
	Object     *serializedObjectExpr     `json:"object,omitempty"`
	Access     *serializedAccessExpr     `json:"access,omitempty"`
	Index      *serializedIndexExpr      `json:"index,omitempty"`
	Call       *serializedCallExpr       `json:"call,omitempty"`
	Unary      *serializedUnaryExpr      `json:"unary,omitempty"`
	Binary     *serializedBinaryExpr     `json:"binary,omitempty"`
	Paren      *serializedParenExpr      `json:"paren,omitempty"`
}

type serializedIdentifierExpr struct {
	Ident *serializedIdent `json:"ident"`
}

type serializedLiteralExpr struct {
	Token    string             `json:"token"`
	Value    string             `json:"value"`
	ValuePos serializedPosition `json:"valuePos"`
}

type serializedArrayExpr struct {
	Elements []serializedExpr   `json:"elements"`
	LBrack   serializedPosition `json:"lBrackPos"`
	RBrack   serializedPosition `json:"rBrackPos"`
}

type serializedObjectExpr struct {
	Fields []serializedObjectField `json:"fields"`
	LCurly serializedPosition      `json:"lCurlyPos"`
	RCurly serializedPosition      `json:"rCurlyPos"`
}

type serializedObjectField struct {
	Name   *serializedIdent `json:"name"`
	Quoted bool             `json:"quoted"`
	Value  *serializedExpr  `json:"value"`
}

type serializedAccessExpr struct {
	Value *serializedExpr  `json:"value"`
	Name  *serializedIdent `json:"name"`
}

type serializedIndexExpr struct {
	Value  *serializedExpr    `json:"value"`
	Index  *serializedExpr    `json:"index"`
	LBrack serializedPosition `json:"lBrackPos"`
	RBrack serializedPosition `json:"rBrackPos"`
}

type serializedCallExpr struct {
	Value  *serializedExpr    `json:"value"`
	Args   []serializedExpr   `json:"args"`
	LParen serializedPosition `json:"lParenPos"`
	RParen serializedPosition `json:"rParenPos"`
}

type serializedUnaryExpr struct {
	Operator    string             `json:"operator"`
	OperatorPos serializedPosition `json:"operatorPos"`
	Value       *serializedExpr    `json:"value"`
}

type serializedBinaryExpr struct {
	Operator    string             `json:"operator"`
	OperatorPos serializedPosition `json:"operatorPos"`
	Left        *serializedExpr    `json:"left"`
	Right       *serializedExpr    `json:"right"`
}

type serializedParenExpr struct {
	Inner  *serializedExpr    `json:"inner"`
	LParen serializedPosition `json:"lParenPos"`
	RParen serializedPosition `json:"rParenPos"`
}

type serializedIdent struct {
	Name string             `json:"name"`
	Pos  serializedPosition `json:"pos"`
}

type serializedDiagnostic struct {
	Severity string             `json:"severity"`
	Message  string             `json:"message"`
	Value    string             `json:"value,omitempty"`
	Start    serializedPosition `json:"start"`
	End      serializedPosition `json:"end"`
}

type serializedPosition struct {
	Filename string `json:"filename,omitempty"`
	Offset   int    `json:"offset"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Valid    bool   `json:"valid"`
}

func main() {}

//export AlloyParseFileJSON
func AlloyParseFileJSON(filename *C.char, data *C.char, dataLen C.int) C.AlloyJSONResult {
	goFilename := cStringToGo(filename)
	goData := cBytesToGo(data, dataLen)

	payload, status := parseFileToJSON(goFilename, goData)
	return buildCResult(payload, status)
}

//export AlloyParseExpressionJSON
func AlloyParseExpressionJSON(data *C.char, dataLen C.int) C.AlloyJSONResult {
	goData := cBytesToGo(data, dataLen)

	payload, status := parseExpressionToJSON(goData)
	return buildCResult(payload, status)
}

//export AlloyFree
func AlloyFree(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}

func buildCResult(payload []byte, status resultStatus) C.AlloyJSONResult {
	var res C.AlloyJSONResult
	res.status = C.int(status)

	if len(payload) == 0 {
		return res
	}

	mem := C.CBytes(payload)
	res.data = (*C.char)(mem)
	res.length = C.size_t(len(payload))
	return res
}

func cStringToGo(str *C.char) string {
	if str == nil {
		return ""
	}
	return C.GoString(str)
}

func cBytesToGo(data *C.char, length C.int) []byte {
	if data == nil || length == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(data), length)
}

func parseFileToJSON(filename string, data []byte) ([]byte, resultStatus) {
	file, diags := parseFileWithDiagnostics(filename, data)
	return serializeResult(file, nil, diags)
}

func parseExpressionToJSON(data []byte) ([]byte, resultStatus) {
	expr, diags := parseExpressionWithDiagnostics(string(data))
	return serializeResult(nil, expr, diags)
}

func parseFileWithDiagnostics(filename string, data []byte) (*ast.File, diag.Diagnostics) {
	file, err := parser.ParseFile(filename, data)
	if err == nil {
		return file, nil
	}

	var diags diag.Diagnostics
	if errors.As(err, &diags) {
		return nil, diags
	}

	return nil, diag.Diagnostics{makeInternalDiagnostic(err)}
}

func parseExpressionWithDiagnostics(expr string) (ast.Expr, diag.Diagnostics) {
	value, err := parser.ParseExpression(expr)
	if err == nil {
		return value, nil
	}

	var diags diag.Diagnostics
	if errors.As(err, &diags) {
		return nil, diags
	}

	return nil, diag.Diagnostics{makeInternalDiagnostic(err)}
}

func serializeResult(file *ast.File, expr ast.Expr, diags diag.Diagnostics) ([]byte, resultStatus) {
	serializedDiagnostics := convertDiagnostics(diags)
	if serializedDiagnostics == nil {
		serializedDiagnostics = []serializedDiagnostic{}
	}

	result := parseResultEnvelope{
		SchemaVersion: schemaVersion,
		Diagnostics:   serializedDiagnostics,
	}

	if file != nil {
		result.File = convertFile(file)
	}
	if expr != nil {
		result.Expression = convertExpr(expr)
	}

	status := statusFromDiagnostics(result.Diagnostics)

	payload, err := json.Marshal(result)
	if err != nil {
		fallback := parseResultEnvelope{
			SchemaVersion: schemaVersion,
			Diagnostics: []serializedDiagnostic{
				makeSerializedInternalDiagnostic(err),
			},
		}
		payload, _ = json.Marshal(fallback)
		status = statusInternalError
	}

	return payload, status
}

func convertFile(file *ast.File) *serializedFile {
	if file == nil {
		return nil
	}

	return &serializedFile{
		Name:     file.Name,
		Body:     convertBody(file.Body),
		Comments: convertCommentGroups(file.Comments),
	}
}

func convertBody(body ast.Body) []serializedStmt {
	if len(body) == 0 {
		return []serializedStmt{}
	}

	stmts := make([]serializedStmt, 0, len(body))
	for _, stmt := range body {
		stmts = append(stmts, convertStmt(stmt))
	}
	return stmts
}

func convertStmt(stmt ast.Stmt) serializedStmt {
	switch s := stmt.(type) {
	case *ast.AttributeStmt:
		return serializedStmt{
			Kind: "Attribute",
			Attribute: &serializedAttributeStmt{
				Name:  convertIdent(s.Name),
				Value: convertExpr(s.Value),
			},
		}
	case *ast.BlockStmt:
		return serializedStmt{
			Kind: "Block",
			Block: &serializedBlockStmt{
				Name:      append([]string(nil), s.Name...),
				NamePos:   convertPos(s.NamePos),
				Label:     s.Label,
				LabelPos:  convertPos(s.LabelPos),
				Body:      convertBody(s.Body),
				LCurlyPos: convertPos(s.LCurlyPos),
				RCurlyPos: convertPos(s.RCurlyPos),
			},
		}
	default:
		return serializedStmt{Kind: "Unknown"}
	}
}

func convertExpr(expr ast.Expr) *serializedExpr {
	if expr == nil {
		return nil
	}

	result := &serializedExpr{
		Secret: expr.IsSecret(),
	}

	switch e := expr.(type) {
	case *ast.IdentifierExpr:
		result.Kind = "Identifier"
		result.Identifier = &serializedIdentifierExpr{
			Ident: convertIdent(e.Ident),
		}
	case *ast.LiteralExpr:
		result.Kind = "Literal"
		result.Literal = &serializedLiteralExpr{
			Token:    e.Kind.String(),
			Value:    e.Value,
			ValuePos: convertPos(e.ValuePos),
		}
	case *ast.ArrayExpr:
		result.Kind = "Array"
		result.Array = &serializedArrayExpr{
			Elements: convertExprList(e.Elements),
			LBrack:   convertPos(e.LBrackPos),
			RBrack:   convertPos(e.RBrackPos),
		}
	case *ast.ObjectExpr:
		result.Kind = "Object"
		result.Object = &serializedObjectExpr{
			Fields: convertObjectFields(e.Fields),
			LCurly: convertPos(e.LCurlyPos),
			RCurly: convertPos(e.RCurlyPos),
		}
	case *ast.AccessExpr:
		result.Kind = "Access"
		result.Access = &serializedAccessExpr{
			Value: convertExpr(e.Value),
			Name:  convertIdent(e.Name),
		}
	case *ast.IndexExpr:
		result.Kind = "Index"
		result.Index = &serializedIndexExpr{
			Value:  convertExpr(e.Value),
			Index:  convertExpr(e.Index),
			LBrack: convertPos(e.LBrackPos),
			RBrack: convertPos(e.RBrackPos),
		}
	case *ast.CallExpr:
		result.Kind = "Call"
		result.Call = &serializedCallExpr{
			Value:  convertExpr(e.Value),
			Args:   convertExprList(e.Args),
			LParen: convertPos(e.LParenPos),
			RParen: convertPos(e.RParenPos),
		}
	case *ast.UnaryExpr:
		result.Kind = "Unary"
		result.Unary = &serializedUnaryExpr{
			Operator:    e.Kind.String(),
			OperatorPos: convertPos(e.KindPos),
			Value:       convertExpr(e.Value),
		}
	case *ast.BinaryExpr:
		result.Kind = "Binary"
		result.Binary = &serializedBinaryExpr{
			Operator:    e.Kind.String(),
			OperatorPos: convertPos(e.KindPos),
			Left:        convertExpr(e.Left),
			Right:       convertExpr(e.Right),
		}
	case *ast.ParenExpr:
		result.Kind = "Paren"
		result.Paren = &serializedParenExpr{
			Inner:  convertExpr(e.Inner),
			LParen: convertPos(e.LParenPos),
			RParen: convertPos(e.RParenPos),
		}
	default:
		result.Kind = "Unknown"
	}

	return result
}

func convertExprList(list []ast.Expr) []serializedExpr {
	if len(list) == 0 {
		return []serializedExpr{}
	}

	out := make([]serializedExpr, 0, len(list))
	for _, expr := range list {
		if serialized := convertExpr(expr); serialized != nil {
			out = append(out, *serialized)
		}
	}
	return out
}

func convertObjectFields(fields []*ast.ObjectField) []serializedObjectField {
	if len(fields) == 0 {
		return []serializedObjectField{}
	}

	out := make([]serializedObjectField, 0, len(fields))
	for _, field := range fields {
		if field == nil {
			continue
		}
		out = append(out, serializedObjectField{
			Name:   convertIdent(field.Name),
			Quoted: field.Quoted,
			Value:  convertExpr(field.Value),
		})
	}
	return out
}

func convertIdent(ident *ast.Ident) *serializedIdent {
	if ident == nil {
		return nil
	}

	return &serializedIdent{
		Name: ident.Name,
		Pos:  convertPos(ident.NamePos),
	}
}

func convertCommentGroups(groups []ast.CommentGroup) []serializedCommentGroup {
	if len(groups) == 0 {
		return []serializedCommentGroup{}
	}

	out := make([]serializedCommentGroup, 0, len(groups))
	for _, group := range groups {
		if len(group) == 0 {
			out = append(out, serializedCommentGroup{})
			continue
		}

		comments := make([]serializedComment, 0, len(group))
		for _, comment := range group {
			if comment == nil {
				continue
			}
			comments = append(comments, serializedComment{
				Text:  comment.Text,
				Start: convertPos(comment.StartPos),
			})
		}
		out = append(out, serializedCommentGroup{Comments: comments})
	}
	return out
}

func convertPos(pos token.Pos) serializedPosition {
	if !pos.Valid() || pos.File() == nil {
		return serializedPosition{}
	}
	return convertTokenPosition(pos.Position())
}

func convertTokenPosition(pos token.Position) serializedPosition {
	return serializedPosition{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
		Valid:    pos.Valid(),
	}
}

func convertDiagnostics(diags diag.Diagnostics) []serializedDiagnostic {
	if len(diags) == 0 {
		return nil
	}

	out := make([]serializedDiagnostic, 0, len(diags))
	for _, d := range diags {
		out = append(out, serializedDiagnostic{
			Severity: severityToString(d.Severity),
			Message:  d.Message,
			Value:    d.Value,
			Start:    convertTokenPosition(d.StartPos),
			End:      convertTokenPosition(d.EndPos),
		})
	}
	return out
}

func severityToString(sev diag.Severity) string {
	switch sev {
	case diag.SeverityLevelWarn:
		return "warning"
	case diag.SeverityLevelError:
		return "error"
	default:
		return "unknown"
	}
}

func statusFromDiagnostics(diags []serializedDiagnostic) resultStatus {
	hasWarn := false
	hasError := false

	for _, d := range diags {
		switch d.Severity {
		case "error":
			hasError = true
		case "warning":
			hasWarn = true
		}
	}

	switch {
	case hasError:
		return statusError
	case hasWarn:
		return statusWarn
	default:
		return statusOK
	}
}

func makeInternalDiagnostic(err error) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.SeverityLevelError,
		Message:  "internal error: " + err.Error(),
	}
}

func makeSerializedInternalDiagnostic(err error) serializedDiagnostic {
	return serializedDiagnostic{
		Severity: "error",
		Message:  "internal error: " + err.Error(),
	}
}
