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

// All textext nodes implement the Text interface.
type Text interface {
	Node
	textNode()
}

// ----------------------------------------------------------------------------
// Expressions

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

// Pos and End implementations for expression/type nodes.

func (x *Ident) Pos() template.Pos      { return x.NamePos }
func (x *BasicLit) Pos() template.Pos   { return x.ValuePos }
func (x *ParenExpr) Pos() template.Pos  { return x.Lparen }
func (x *IndexExpr) Pos() template.Pos  { return x.X.Pos() }
func (x *CallExpr) Pos() template.Pos   { return x.Fun.Pos() }
func (x *BinaryExpr) Pos() template.Pos { return x.X.Pos() }

func (x *Ident) End() template.Pos      { return template.Pos(int(x.NamePos) + len(x.Name)) }
func (x *BasicLit) End() template.Pos   { return template.Pos(int(x.ValuePos) + len(x.Value)) }
func (x *ParenExpr) End() template.Pos  { return x.Rparen + 1 }
func (x *IndexExpr) End() template.Pos  { return x.Rbrack + 1 }
func (x *CallExpr) End() template.Pos   { return x.Rparen + 1 }
func (x *BinaryExpr) End() template.Pos { return x.Y.End() }

// exprNode() ensures that only expression/type nodes can be
// assigned to an Expr.
//
func (*Ident) exprNode()      {}
func (*BasicLit) exprNode()   {}
func (*ParenExpr) exprNode()  {}
func (*IndexExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*BinaryExpr) exprNode() {}

// ----------------------------------------------------------------------------
// Statements

// A statement is represented by a tree consisting of one
// or more of the following concrete statement nodes.
//
type (
	// An AssignStmt node represents an assignment or
	// a short variable declaration.
	//
	AssignStmt struct {
		Lhs    []Expr
		TokPos template.Pos   // position of Tok
		Tok    template.Token // assignment token, DEFINE
		Rhs    []Expr
	}

	// An IfStmt node represents an if statement.
	IfStmt struct {
		If   template.Pos // position of "if" keyword
		Init Stmt         // initialization statement; or nil
		Cond Expr         // condition
		Else Stmt         // else branch; or nil
	}

	// A ForStmt represents a for statement.
	ForStmt struct {
		For  template.Pos // position of "for" keyword
		Init Stmt         // initialization statement; or nil
		Cond Expr         // condition; or nil
		Post Stmt         // post iteration statement; or nil
	}
)
