package main

import (
	"github.com/grafana/alloy/syntax/ast"
	"github.com/grafana/alloy/syntax/diag"
	"github.com/grafana/alloy/syntax/token"
)

type serializedFile struct {
	Name     string                   `json:"name"`
	Body     []serializedStmt         `json:"body"`
	Comments []serializedCommentGroup `json:"comments,omitempty"`
}

type serializedStmt struct {
	Kind      string               `json:"kind"`
	Attribute *serializedAttribute `json:"attribute,omitempty"`
	Block     *serializedBlock     `json:"block,omitempty"`
}

type serializedAttribute struct {
	Name  serializedIdent `json:"name"`
	Value *serializedExpr `json:"value"`
}

type serializedBlock struct {
	Name      []string         `json:"name"`
	NamePos   serializedPos    `json:"namePos"`
	Label     string           `json:"label,omitempty"`
	LabelPos  serializedPos    `json:"labelPos"`
	Body      []serializedStmt `json:"body"`
	LCurlyPos serializedPos    `json:"lCurlyPos"`
	RCurlyPos serializedPos    `json:"rCurlyPos"`
}

type serializedExpr struct {
	Kind       string                    `json:"kind"`
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
	Ident  serializedIdent `json:"ident"`
	Secret bool            `json:"secret"`
}

type serializedLiteralExpr struct {
	Token    string        `json:"token"`
	Value    string        `json:"value"`
	ValuePos serializedPos `json:"valuePos"`
	Secret   bool          `json:"secret"`
}

type serializedArrayExpr struct {
	Elements  []*serializedExpr `json:"elements"`
	LBrackPos serializedPos     `json:"lBrackPos"`
	RBrackPos serializedPos     `json:"rBrackPos"`
	Secret    bool              `json:"secret"`
}

type serializedObjectExpr struct {
	Fields    []serializedObjectField `json:"fields"`
	LCurlyPos serializedPos           `json:"lCurlyPos"`
	RCurlyPos serializedPos           `json:"rCurlyPos"`
	Secret    bool                    `json:"secret"`
}

type serializedObjectField struct {
	Name   *serializedIdent `json:"name,omitempty"`
	Quoted bool             `json:"quoted"`
	Value  *serializedExpr  `json:"value"`
}

type serializedAccessExpr struct {
	Value  *serializedExpr `json:"value"`
	Name   serializedIdent `json:"name"`
	Secret bool            `json:"secret"`
}

type serializedIndexExpr struct {
	Value     *serializedExpr `json:"value"`
	Index     *serializedExpr `json:"index"`
	LBrackPos serializedPos   `json:"lBrackPos"`
	RBrackPos serializedPos   `json:"rBrackPos"`
	Secret    bool            `json:"secret"`
}

type serializedCallExpr struct {
	Value     *serializedExpr   `json:"value"`
	Args      []*serializedExpr `json:"args"`
	LParenPos serializedPos     `json:"lParenPos"`
	RParenPos serializedPos     `json:"rParenPos"`
	Secret    bool              `json:"secret"`
}

type serializedUnaryExpr struct {
	Token    string          `json:"token"`
	TokenPos serializedPos   `json:"tokenPos"`
	Value    *serializedExpr `json:"value"`
	Secret   bool            `json:"secret"`
}

type serializedBinaryExpr struct {
	Token    string          `json:"token"`
	TokenPos serializedPos   `json:"tokenPos"`
	Left     *serializedExpr `json:"left"`
	Right    *serializedExpr `json:"right"`
	Secret   bool            `json:"secret"`
}

type serializedParenExpr struct {
	Inner     *serializedExpr `json:"inner"`
	LParenPos serializedPos   `json:"lParenPos"`
	RParenPos serializedPos   `json:"rParenPos"`
	Secret    bool            `json:"secret"`
}

type serializedIdent struct {
	Name string        `json:"name"`
	Pos  serializedPos `json:"pos"`
}

type serializedCommentGroup struct {
	Comments []serializedComment `json:"comments"`
}

type serializedComment struct {
	Text string        `json:"text"`
	Pos  serializedPos `json:"pos"`
}

type serializedDiagnostic struct {
	Severity string           `json:"severity"`
	Message  string           `json:"message"`
	Value    string           `json:"value,omitempty"`
	Start    serializedSource `json:"start"`
	End      serializedSource `json:"end,omitempty"`
}

type serializedSource struct {
	Filename string `json:"filename"`
	Offset   int    `json:"offset"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

type serializedPos struct {
	Filename string `json:"filename"`
	Offset   int    `json:"offset"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

func serializeFile(f *ast.File) serializedFile {
	result := serializedFile{
		Name: f.Name,
		Body: serializeBody(f.Body),
	}

	if len(f.Comments) > 0 {
		result.Comments = make([]serializedCommentGroup, 0, len(f.Comments))
		for _, group := range f.Comments {
			result.Comments = append(result.Comments, serializeCommentGroup(group))
		}
	}
	return result
}

func serializeBody(body ast.Body) []serializedStmt {
	if len(body) == 0 {
		return nil
	}

	out := make([]serializedStmt, 0, len(body))
	for _, stmt := range body {
		out = append(out, serializeStmt(stmt))
	}
	return out
}

func serializeStmt(stmt ast.Stmt) serializedStmt {
	switch s := stmt.(type) {
	case *ast.AttributeStmt:
		return serializedStmt{
			Kind: "attribute",
			Attribute: &serializedAttribute{
				Name:  serializeIdent(s.Name),
				Value: serializeExpr(s.Value),
			},
		}
	case *ast.BlockStmt:
		return serializedStmt{
			Kind: "block",
			Block: &serializedBlock{
				Name:      append([]string(nil), s.Name...),
				NamePos:   serializePos(s.NamePos),
				Label:     s.Label,
				LabelPos:  serializePos(s.LabelPos),
				Body:      serializeBody(s.Body),
				LCurlyPos: serializePos(s.LCurlyPos),
				RCurlyPos: serializePos(s.RCurlyPos),
			},
		}
	default:
		return serializedStmt{Kind: "unknown"}
	}
}

func serializeExpr(expr ast.Expr) *serializedExpr {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.IdentifierExpr:
		return &serializedExpr{
			Kind: "identifier",
			Identifier: &serializedIdentifierExpr{
				Ident:  serializeIdent(e.Ident),
				Secret: e.Secret,
			},
		}
	case *ast.LiteralExpr:
		return &serializedExpr{
			Kind: "literal",
			Literal: &serializedLiteralExpr{
				Token:    e.Kind.String(),
				Value:    e.Value,
				ValuePos: serializePos(e.ValuePos),
				Secret:   e.Secret,
			},
		}
	case *ast.ArrayExpr:
		elements := make([]*serializedExpr, 0, len(e.Elements))
		for _, el := range e.Elements {
			elements = append(elements, serializeExpr(el))
		}
		return &serializedExpr{
			Kind: "array",
			Array: &serializedArrayExpr{
				Elements:  elements,
				LBrackPos: serializePos(e.LBrackPos),
				RBrackPos: serializePos(e.RBrackPos),
				Secret:    e.Secret,
			},
		}
	case *ast.ObjectExpr:
		fields := make([]serializedObjectField, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, serializeObjectField(field))
		}
		return &serializedExpr{
			Kind: "object",
			Object: &serializedObjectExpr{
				Fields:    fields,
				LCurlyPos: serializePos(e.LCurlyPos),
				RCurlyPos: serializePos(e.RCurlyPos),
				Secret:    e.Secret,
			},
		}
	case *ast.AccessExpr:
		return &serializedExpr{
			Kind: "access",
			Access: &serializedAccessExpr{
				Value:  serializeExpr(e.Value),
				Name:   serializeIdent(e.Name),
				Secret: e.Secret,
			},
		}
	case *ast.IndexExpr:
		return &serializedExpr{
			Kind: "index",
			Index: &serializedIndexExpr{
				Value:     serializeExpr(e.Value),
				Index:     serializeExpr(e.Index),
				LBrackPos: serializePos(e.LBrackPos),
				RBrackPos: serializePos(e.RBrackPos),
				Secret:    e.Secret,
			},
		}
	case *ast.CallExpr:
		args := make([]*serializedExpr, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, serializeExpr(arg))
		}
		return &serializedExpr{
			Kind: "call",
			Call: &serializedCallExpr{
				Value:     serializeExpr(e.Value),
				Args:      args,
				LParenPos: serializePos(e.LParenPos),
				RParenPos: serializePos(e.RParenPos),
				Secret:    e.Secret,
			},
		}
	case *ast.UnaryExpr:
		return &serializedExpr{
			Kind: "unary",
			Unary: &serializedUnaryExpr{
				Token:    e.Kind.String(),
				TokenPos: serializePos(e.KindPos),
				Value:    serializeExpr(e.Value),
				Secret:   e.Secret,
			},
		}
	case *ast.BinaryExpr:
		return &serializedExpr{
			Kind: "binary",
			Binary: &serializedBinaryExpr{
				Token:    e.Kind.String(),
				TokenPos: serializePos(e.KindPos),
				Left:     serializeExpr(e.Left),
				Right:    serializeExpr(e.Right),
				Secret:   e.Secret,
			},
		}
	case *ast.ParenExpr:
		return &serializedExpr{
			Kind: "paren",
			Paren: &serializedParenExpr{
				Inner:     serializeExpr(e.Inner),
				LParenPos: serializePos(e.LParenPos),
				RParenPos: serializePos(e.RParenPos),
				Secret:    e.Secret,
			},
		}
	default:
		return &serializedExpr{Kind: "unknown"}
	}
}

func serializeObjectField(field *ast.ObjectField) serializedObjectField {
	var name *serializedIdent
	if field.Name != nil {
		ident := serializeIdent(field.Name)
		name = &ident
	}
	return serializedObjectField{
		Name:   name,
		Quoted: field.Quoted,
		Value:  serializeExpr(field.Value),
	}
}

func serializeIdent(ident *ast.Ident) serializedIdent {
	if ident == nil {
		return serializedIdent{}
	}
	return serializedIdent{
		Name: ident.Name,
		Pos:  serializePos(ident.NamePos),
	}
}

func serializeCommentGroup(group ast.CommentGroup) serializedCommentGroup {
	comments := make([]serializedComment, 0, len(group))
	for _, comment := range group {
		comments = append(comments, serializedComment{
			Text: comment.Text,
			Pos:  serializePos(comment.StartPos),
		})
	}
	return serializedCommentGroup{Comments: comments}
}

func serializeDiagnostics(diags diag.Diagnostics) []serializedDiagnostic {
	if len(diags) == 0 {
		return []serializedDiagnostic{}
	}
	out := make([]serializedDiagnostic, 0, len(diags))
	for _, d := range diags {
		out = append(out, serializeDiagnostic(d))
	}
	return out
}

func serializeDiagnostic(d diag.Diagnostic) serializedDiagnostic {
	return serializedDiagnostic{
		Severity: serializeSeverity(d.Severity),
		Message:  d.Message,
		Value:    d.Value,
		Start:    serializeSource(d.StartPos),
		End:      serializeSource(d.EndPos),
	}
}

func serializeSeverity(sev diag.Severity) string {
	switch sev {
	case diag.SeverityLevelWarn:
		return "warning"
	case diag.SeverityLevelError:
		return "error"
	default:
		return "unknown"
	}
}

func serializeSource(pos token.Position) serializedSource {
	if !pos.Valid() {
		return serializedSource{}
	}
	return serializedSource{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

func serializePos(pos token.Pos) serializedPos {
	if !pos.Valid() || pos.File() == nil {
		return serializedPos{}
	}

	p := pos.Position()
	return serializedPos{
		Filename: p.Filename,
		Offset:   p.Offset,
		Line:     p.Line,
		Column:   p.Column,
	}
}
