package template

const NoPos Pos = 0

type Pos int

func (p Pos) Position() Pos {
	return p
}

type ASTNode interface {
	Pos() Pos // position of first character belonging to the node
	End() Pos // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	ASTNode
	exprNode()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	ASTNode
	stmtNode()
}

// All textext nodes implement the Text interface.
type Text interface {
	ASTNode
	textNode()
}

// ----------------------------------------------------------------------------
// Expressions

type (
	Ident struct {
		NamePos Pos    // identifier position
		Name    string // identifier name
	}

	BasicLit struct {
		ValuePos Pos    // literal position
		Kind     int    // TYPE_NUMBER, TYPE_STRING
		Value    string // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', etc.
	}

	OpLit struct {
		OpPos Pos    // literal position
		Op    string // literal string; e.g. + - * /
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X     Expr // expression
		Index Expr // index expression
	}

	ParenExpr struct {
		Kind  int    // TYPE_OPERATOR
		Paren string // literal paren; (, )
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun    Expr      // function expression
		Lparen Pos       // position of "("
		Args   *ArgsExpr // function arguments; or nil
		Rparen Pos       // position of ")"
	}

	ArgsExpr struct {
		List []Expr // function arguments
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X  Expr  // left operand
		Op OpLit // operator
		Y  Expr  // right operand
	}
)

// Pos and End implementations for expression/type nodes.

func (x *Ident) Pos() Pos     { return x.NamePos }
func (x *BasicLit) Pos() Pos  { return x.ValuePos }
func (x *OpLit) Pos() Pos     { return x.OpPos }
func (x *IndexExpr) Pos() Pos { return x.X.Pos() }
func (x *CallExpr) Pos() Pos  { return x.Fun.Pos() }
func (x *ArgsExpr) Pos() Pos {
	if len(x.List) > 0 {
		return x.List[0].Pos()
	}
	return NoPos
}
func (x *BinaryExpr) Pos() Pos { return x.X.Pos() }

func (x *Ident) End() Pos     { return Pos(int(x.NamePos) + len(x.Name)) }
func (x *BasicLit) End() Pos  { return Pos(int(x.ValuePos) + len(x.Value)) }
func (x *OpLit) End() Pos     { return Pos(int(x.OpPos) + len(x.Op)) }
func (x *IndexExpr) End() Pos { return x.Index.End() + 2 }
func (x *CallExpr) End() Pos  { return x.Rparen + 1 }
func (x *ArgsExpr) End() Pos {
	if len(x.List) > 0 {
		return x.List[len(x.List)-1].Pos()
	}
	return NoPos
}
func (x *BinaryExpr) End() Pos { return x.Y.End() }

// exprNode() ensures that only expression/type nodes can be
// assigned to an Expr.
//
func (*Ident) exprNode()      {}
func (*BasicLit) exprNode()   {}
func (*OpLit) exprNode()      {}
func (*IndexExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*ArgsExpr) exprNode()   {}
func (*BinaryExpr) exprNode() {}

// ----------------------------------------------------------------------------
// Convenience functions for Idents

// NewIdent creates a new Ident without position.
// Useful for ASTs generated by code other than the Go parser.
//
func NewIdent(name string) *Ident { return &Ident{NoPos, name} }

// ----------------------------------------------------------------------------
// Statements

type (
	// TextStmt
	TextStmt struct {
		Text Expr // text content BasicLit
	}

	ValueStmt struct {
		Tok Expr // assignment expr
	}

	// An AssignStmt node represents an assignment or
	// a short variable declaration.
	//
	AssignStmt struct {
		Lh     Expr   // Ident
		TokPos Pos    // position of Tok
		Tok    string // assignment token, DEFINE
		Rh     Expr
	}

	// A SectionStmt node represents a braced statement list.
	SectionStmt struct {
		List []Stmt
	}

	// An IfStmt node represents an if statement.
	IfStmt struct {
		If   Pos  // position of "if" keyword
		Cond Expr // condition
		Else Stmt // else branch; or nil
		Body *SectionStmt
	}

	// A ForStmt represents a for statement.
	ForStmt struct {
		For  Pos  // position of "for" keyword
		Init Stmt // initialization statement; or nil
		Cond Expr // condition; or nil
		Post Stmt // post iteration statement; or nil
		Body *SectionStmt
	}

	// A RangeStmt represents a for statement with a range clause.
	RangeStmt struct {
		For        Pos    // position of "for" keyword
		Key, Value Expr   // Key, Value may be nil
		TokPos     Pos    // position of Tok; invalid if Key == nil
		Tok        string // ILLEGAL if Key == nil, ASSIGN, DEFINE
		X          Expr   // value to range over
		Body       *SectionStmt
	}

	//
	BlockStmt struct {
		Name string
		Body *SectionStmt
	}
)

// Pos and End implementations for statement nodes.
func (s *TextStmt) Pos() Pos   { return s.Text.Pos() }
func (s *ValueStmt) Pos() Pos  { return s.Tok.Pos() }
func (s *AssignStmt) Pos() Pos { return s.Lh.Pos() }
func (s *SectionStmt) Pos() Pos {
	if len(s.List) > 0 {
		return s.List[0].Pos()
	}
	return NoPos
}
func (s *IfStmt) Pos() Pos     { return s.If }
func (s *ForStmt) Pos() Pos    { return s.For }
func (s *RangeStmt) Pos() Pos  { return s.For }
func (s *BlockStmt) Pos() Pos  { return s.Body.Pos() }
func (s *TextStmt) End() Pos   { return s.Text.End() }
func (s *ValueStmt) End() Pos  { return s.Tok.End() }
func (s *AssignStmt) End() Pos { return s.Rh.End() }
func (s *SectionStmt) End() Pos {
	if len(s.List) > 0 {
		return s.List[len(s.List)-1].Pos()
	}
	return NoPos
}
func (s *IfStmt) End() Pos {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}
func (s *ForStmt) End() Pos   { return s.Body.End() }
func (s *RangeStmt) End() Pos { return s.Body.End() }
func (s *BlockStmt) End() Pos { return s.Body.End() }

// stmtNode() ensures that only statement nodes can be
// assigned to a Stmt.
//

func (*TextStmt) stmtNode()    {}
func (*ValueStmt) stmtNode()   {}
func (*AssignStmt) stmtNode()  {}
func (*SectionStmt) stmtNode() {}
func (*IfStmt) stmtNode()      {}
func (*ForStmt) stmtNode()     {}
func (*RangeStmt) stmtNode()   {}
func (*BlockStmt) stmtNode()   {}
