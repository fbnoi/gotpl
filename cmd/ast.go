package main

import (
	"gotpl/template"
)

type Node interface {
	Pos() template.Pos // position of first character belonging to the node
	End() template.Pos // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node
	stmtNode()
}

// All statement nodes implement the Stmt interface.
type Text interface {
	Node
	textNode()
}

type (
	Ident struct {
		NamePos template.Pos // identifier position
		Name    string       // identifier name
	}

	BasicLit struct {
		ValuePos template.Pos // literal position
		Kind     int          // template.TYPE_NUMBER, template.TYPE_STRING
		Value    string       // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		Lparen template.Pos // position of "("
		X      Expr         // parenthesized expression
		Rparen template.Pos // position of ")"
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X      Expr         // expression
		Lbrack template.Pos // position of "["
		Index  Expr         // index expression
		Rbrack template.Pos // position of "]"
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun    Expr         // function expression
		Lparen template.Pos // position of "("
		Args   []Expr       // function arguments; or nil
		Rparen template.Pos // position of ")"
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X     Expr           // left operand
		OpPos template.Pos   // position of Op
		Op    template.Token // operator
		Y     Expr           // right operand
	}
)
