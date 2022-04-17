package template

import (
	"errors"
	"fmt"
)

var opPriority = map[string]int{
	"==": 0, ">=": 0, "<=": 0, ">": 0, "<": 0, "!=": 0,
	"+": 5, "-": 5, "%": 10, "*": 10, "/": 10, "[": 15,
}

type Tree struct {
	List   []ASTNode
	Extend *ExtendStmt
}

type TokenFilter struct {
	*TokenStream
	Tr     *Tree
	Cursor Stmt
	Stack  []Stmt
}

func (filter *TokenFilter) Filter(stream *TokenStream) (*Tree, error) {
	filter.TokenStream = stream
	var err error
	for !stream.IsEOF() {
		token := filter.Next()
		switch token.Type() {
		case TYPE_TEXT:
			filter.parseText()
		case TYPE_VAR_START:
			err = filter.parseVar()
		case TYPE_BLOCK_START:
			token := filter.Next()
			switch token.Value() {
			case "if":
				err = filter.parseIf()
			case "else":
				err = filter.parseElse()
			case "elseif":
				err = filter.parseElseIf()
			case "endif":
				err = filter.popIf()
			case "for":
				err = filter.parseFor()
			case "endfor":
				err = filter.popFor()
			case "range":
				err = filter.parseRange()
			case "endrange":
				err = filter.popRange()
			case "block":
				err = filter.parseBlock()
			case "endblock":
				err = filter.popBlock()
			case "set":
				err = filter.parseSet()
			case "include":
				err = filter.parseInclude()
			case "extend":
				err = filter.parseExtend(token)
			default:
				return nil, filter.unexpected(token)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return filter.Tr, nil
}

func (filter *TokenFilter) parseExtend(token *Token) error {
	if filter.Tr.Extend != nil {
		return filter.unexpected(token)
	}
	es := &ExtendStmt{}
	if token := filter.Next(); token.Type() == TYPE_STRING {
		es.Ident = &BasicLit{
			Kind:  TYPE_STRING,
			Value: token.Value(),
		}
		filter.Tr.Extend = es
		return nil
	}
	return filter.unexpected(token)
}

func (filter *TokenFilter) parseInclude() (err error) {
	is := &IncludeStmt{}
	if token := filter.Next(); token.Type() == TYPE_STRING {
		is.Ident = &BasicLit{
			Kind:  TYPE_STRING,
			Value: token.Value(),
		}

		if token = filter.Next(); token.Value() == "with" {
			for !filter.IsEOF() && token.Type() != TYPE_BLOCK_END {
				var ts []*Token
				token = filter.Next()
				for token.Value() != ";" &&
					token.Type() != TYPE_EOF &&
					token.Type() != TYPE_BLOCK_END {
					ts = append(ts, token)
					token = filter.Next()
				}
				if len(ts) > 0 {
					var as *AssignStmt
					if as, err = parseAssignStmt(ts); err != nil {
						return
					} else if as != nil {
						is.Params = append(is.Params, as)
					}
				}
			}
		} else if token.Type() != TYPE_BLOCK_END {
			return filter.unexpected(token)
		}
		filter.append(is)
		return nil
	}
	return filter.unexpected(filter.Current())
}

func (filter *TokenFilter) parseText() {
	t := &TextStmt{&BasicLit{
		Kind:  TYPE_STRING,
		Value: filter.Current().Value(),
	}}
	filter.append(t)
}

func (filter *TokenFilter) parseVar() (err error) {
	vs := &ValueStmt{}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_VAR_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	if vs.Tok, err = parseExpr(ts); err == nil {
		filter.append(vs)
	}
	return
}

func (filter *TokenFilter) parseIf() (err error) {
	is := &IfStmt{}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	if is.Cond, err = parseExpr(ts); err == nil {
		filter.append(is)
		filter.push(is)
	}
	return
}

func (filter *TokenFilter) parseElse() (err error) {
	es := &SectionStmt{}
	if st, ok := filter.Cursor.(*IfStmt); ok {
		st.Else = es
	} else {
		err = filter.unexpected(filter.Current())
	}
	filter.push(es)
	return
}

func (filter *TokenFilter) parseElseIf() (err error) {
	efs := &IfStmt{}
	if st, ok := filter.Cursor.(*IfStmt); ok {
		st.Else = efs
	} else {
		return filter.unexpected(filter.Current())
	}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	if efs.Cond, err = parseExpr(ts); err != nil {
		return
	}
	// elseif do not in stack
	filter.Cursor = efs
	return nil
}

func (filter *TokenFilter) parseFor() (err error) {
	fs := &ForStmt{}
	var (
		tss   [][]*Token
		token *Token
	)
	for !filter.IsEOF() && filter.Current().Type() != TYPE_BLOCK_END {
		var ts []*Token
		token = filter.Next()
		for token.Value() != ";" &&
			token.Type() != TYPE_EOF &&
			token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
			token = filter.Next()
		}
		tss = append(tss, ts)
	}
	if len(tss) == 3 {
		if fs.Init, err = parseAssignStmt(tss[0]); err != nil {
			return
		} else if fs.Cond, err = parseExpr(tss[1]); err != nil {
			return
		} else if fs.Post, err = parseAssignStmt(tss[2]); err != nil {
			return
		}
	} else if len(tss) == 1 {
		if fs.Cond, err = parseExpr(tss[0]); err != nil {
			return
		}
	} else {
		err = filter.unexpected(token)
		return
	}
	filter.append(fs)
	filter.push(fs)
	return
}

func (filter *TokenFilter) parseRange() (err error) {
	rs := &RangeStmt{}
	keyToken := filter.Next()
	rs.Key = &Ident{keyToken.Value()}
	valueToken := filter.Next()
	if valueToken.Value() == "," {
		valueToken = filter.Next()
	} else if valueToken.Value() == "=" {
		valueToken = nil
	} else {
		err = filter.unexpected(valueToken)
		return
	}
	if valueToken != nil && valueToken.Value() != "_" {
		rs.Value = &Ident{valueToken.Value()}
	}
	var (
		ts    []*Token
		token *Token
	)
	for !filter.IsEOF() {
		token = filter.Next()
		if token.Type() == TYPE_BLOCK_END {
			break
		}
		ts = append(ts, token)
	}
	if rs.X, err = parseExpr(ts); err == nil {
		filter.append(rs)
		filter.push(rs)
	}
	return
}

func (filter *TokenFilter) parseBlock() error {
	token := filter.Next()
	if token.Type() != TYPE_NAME {
		return filter.unexpected(token)
	}
	bs := &BlockStmt{
		Name: &Ident{Name: token.Value()},
	}
	filter.append(bs)
	filter.push(bs)
	return nil
}

func (filter *TokenFilter) parseSet() (err error) {
	ss := &SetStmt{}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	if ss.Assign, err = parseAssignStmt(ts); err == nil {
		filter.append(ss)
	}
	return
}

func (filter *TokenFilter) append(s Stmt) error {
	if filter.Cursor == nil {
		filter.Tr.List = append(filter.Tr.List, s)
		return nil
	}
	if st, ok := filter.Cursor.(AppendAble); ok {
		st.Append(s)
		return nil
	}
	return NewParseTemplateFaild(filter.Source, filter.Current().Line())
}

func (filter *TokenFilter) popBlock() (err error) {
	_, ok := filter.Cursor.(*BlockStmt)
	for !ok {
		if filter.Cursor, err = filter.pop(); err != nil {
			return
		}
		_, ok = filter.Cursor.(*BlockStmt)
	}
	filter.Cursor, err = filter.pop()
	return
}

func (filter *TokenFilter) popRange() (err error) {
	_, ok := filter.Cursor.(*RangeStmt)
	for !ok {
		if filter.Cursor, err = filter.pop(); err != nil {
			return
		}
		_, ok = filter.Cursor.(*RangeStmt)
	}
	filter.Cursor, err = filter.pop()
	return
}

func (filter *TokenFilter) popFor() (err error) {
	_, ok := filter.Cursor.(*ForStmt)
	for !ok {
		if filter.Cursor, err = filter.pop(); err != nil {
			return
		}
		_, ok = filter.Cursor.(*ForStmt)
	}
	filter.Cursor, err = filter.pop()
	return
}

func (filter *TokenFilter) popIf() (err error) {
	_, ok := filter.Cursor.(*IfStmt)
	for !ok {
		if filter.Cursor, err = filter.pop(); err != nil {
			return
		}
		_, ok = filter.Cursor.(*IfStmt)
	}
	filter.Cursor, err = filter.pop()
	return
}

func (filter *TokenFilter) pop() (s Stmt, err error) {
	if len(filter.Stack) == 0 {
		if filter.Cursor == nil {
			return nil, errors.New("pop: Try to peek expression from an empty stack")
		}
		return nil, nil
	}
	s = filter.Stack[len(filter.Stack)-1]
	filter.Stack = filter.Stack[:len(filter.Stack)-1]
	return
}

func (filter *TokenFilter) push(s Stmt) {
	if filter.Cursor != nil {
		filter.Stack = append(filter.Stack, filter.Cursor)
	}
	filter.Cursor = s
}

func (filter *TokenFilter) unexpected(token *Token) error {
	return NewUnexpectedToken(filter.Source, token.Line(), token.Value())
}

type ExprWraper struct {
	eStack  []Expr
	opStack []*Token
}

func (ew *ExprWraper) Wrap(stream []*Token) (expr Expr, err error) {
	for i := 0; i < len(stream); i++ {
		token := stream[i]
		switch token.Type() {
		case TYPE_STRING, TYPE_NUMBER:
			ew.pushExpr(&BasicLit{Kind: token.Type(), Value: token.Value()})
		case TYPE_NAME:
			if i+1 < len(stream) {
				p := stream[i+1]
				if p.Value() == "(" {
					ew.pushOp(token)
					continue
				}
			}
			ew.pushExpr(&Ident{token.Value()})
		case TYPE_OPERATOR:
			var ok bool
			switch token.Value() {
			case "+", "-", "*", "/", "%", "==", ">=", "<=", ">", "<", "!=":

				if tt := ew.peekOp(); tt == nil {
					ew.pushOp(token)
				} else if ok, err = comparePriority(token, tt); err != nil {
					return
				} else if ok {
					ew.pushOp(token)
				} else {
					if err = ew.revert(token); err != nil {
						return
					}
				}

			default:
				return nil, fmt.Errorf("Wrap: unexpected operator %s", token.Value())
			}
		case TYPE_PUNCTUATION:
			var op *Token
			switch token.Value() {
			case ",":
				if op = ew.peekOp(); op == nil {
					return nil, fmt.Errorf("Wrap: unexpected punctuation %s", token.Value())
				}
				for op.Value() != "(" && op.Value() != "," {
					if op, err = ew.popOp(); err != nil {
						if err = ew.revert(op); err != nil {
							return
						}
					}
				}
				ew.pushOp(token)
			case "(", "[":
				ew.pushOp(token)
			case "]":
				for {
					if op, err = ew.popOp(); err != nil {
						return
					}
					if op.Value() == "[" {
						break
					}
					if err = ew.revert(op); err != nil {
						return
					}
				}
			case ")":
				for {
					if op, err = ew.popOp(); err != nil {
						return
					}
					if op.Value() == "(" {
						break
					}
					if err = ew.revert(op); err != nil {
						return
					}
				}
				if op = ew.peekOp(); op == nil {
					return nil, fmt.Errorf("Wrap: unexpected punctuation %s", token.Value())
				} else if op.Type() == TYPE_NAME {
					op, _ = ew.popOp()
					if err = ew.revert(op); err != nil {
						return
					}
				}
			default:
				return nil, fmt.Errorf("Wrap: unexpected punctuation %s", token.Value())
			}
		}
	}
	var (
		op *Token
	)
	for len(ew.opStack) > 0 {
		if op, err = ew.popOp(); err != nil {
			return
		}
		if err = ew.revert(op); err != nil {
			return
		}
	}
	expr, err = ew.popExpr()
	if len(ew.eStack) > 0 {
		fmt.Println(ew.eStack)
		panic(len(ew.eStack))
	}
	return
}

func (ew *ExprWraper) revert(op *Token) error {
	if op.Type() == TYPE_NAME {
		fun := &Ident{op.Value()}
		call := &CallExpr{Fun: fun}
		if expr := ew.peekExpr(); expr == nil {
			return fmt.Errorf("revert: unexpected operator %s", op.Value())
		} else if _, ok := expr.(*ArgsExpr); ok {
			args, _ := ew.popExpr()
			call.Args = args.(*ArgsExpr)
		}
		return nil
	}
	e1, err1 := ew.popExpr()
	e2, err2 := ew.popExpr()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	expr, err := waperBinary(op, e1, e2)
	if err != nil {
		return nil
	}
	ew.pushExpr(expr)
	return nil
}

func (ew *ExprWraper) peekExpr() Expr {
	if len(ew.eStack) == 0 {
		return nil
	}
	return ew.eStack[len(ew.eStack)-1]
}

func (ew *ExprWraper) pushExpr(e Expr) {
	ew.eStack = append(ew.eStack, e)
}

func (ew *ExprWraper) popExpr() (Expr, error) {
	if len(ew.eStack) == 0 {
		return nil, errors.New("popExpr: Try to pop expression from an empty stack")
	}
	t := ew.eStack[len(ew.eStack)-1]
	ew.eStack = ew.eStack[:len(ew.eStack)-1]
	return t, nil
}

func (ew *ExprWraper) peekOp() *Token {
	if len(ew.opStack) == 0 {
		return nil
	}
	return ew.opStack[len(ew.opStack)-1]
}

func (ew *ExprWraper) pushOp(op *Token) {
	ew.opStack = append(ew.opStack, op)
}

func (ew *ExprWraper) popOp() (*Token, error) {
	if len(ew.opStack) == 0 {
		return nil, errors.New("peekOp: Try to pop token from an empty stack")
	}
	t := ew.opStack[len(ew.opStack)-1]
	ew.opStack = ew.opStack[:len(ew.opStack)-1]
	return t, nil
}

func waperBinary(op *Token, x1, x2 Expr) (Expr, error) {
	switch op.Type() {
	case TYPE_NAME:
		fn := &Ident{op.Value()}
		return &CallExpr{Fun: fn}, nil
	case TYPE_OPERATOR:
		switch op.Value() {
		case "+", "-", "*", "/", "%", ">", "<", ">=", "<=", "!=", "==":
			return &BinaryExpr{X: x2, Op: OpLit{op.Value()}, Y: x1}, nil
		}
	case TYPE_PUNCTUATION:
		switch op.Value() {
		case "[":
			return &IndexExpr{X: x1, Index: x2}, nil
		case ",":
			if arg, ok := x1.(*ArgsExpr); ok {
				arg.List = append(arg.List, x2)
				return arg, nil
			}
			return &ArgsExpr{List: []Expr{x1, x2}}, nil
		}
	default:

	}
	return nil, fmt.Errorf("waperBinary: unexpected token %s", op.Value())
}

func parseAssignStmt(ts []*Token) (*AssignStmt, error) {
	switch {
	case len(ts) == 0:
		return nil, errors.New("parseAssignStmt: empty token list")
	case len(ts) >= 2:
		token := ts[0]
		if token.Type() != TYPE_NAME {
			return nil, fmt.Errorf("parseAssignStmt: unexpected token %s", token.Value())
		}
		ss := &AssignStmt{Lh: &Ident{token.Value()}}
		tok := ts[1]
		ss.Tok = tok.Value()
		if len(ts) == 2 && (tok.Value() == "++" || tok.Value() == "--") {
			return ss, nil
		} else if tok.Value() == "-=" || tok.Value() == "+=" || tok.Value() == "=" {
			if expr, err := parseExpr(ts[2:]); err == nil {
				ss.Rh = expr
				return ss, nil
			} else {
				return nil, err
			}
		}
	}
	return nil, errors.New("parseAssignStmt: parse failed")
}

func parseExpr(ts []*Token) (Expr, error) {
	return (&ExprWraper{}).Wrap(ts)
}

func comparePriority(t1, t2 *Token) (bool, error) {
	if t1.Type() == TYPE_NAME {
		return true, nil
	}
	if t1.Value() == "(" || t1.Value() == "[" {
		return true, nil
	}

	if t2.Type() == TYPE_NAME {
		return false, nil
	}

	if t2.Value() == "(" || t2.Value() == "[" {
		return false, nil
	}

	if t2.Value() == "," {
		return true, nil
	}

	p1, ok1 := opPriority[t1.Value()]
	p2, ok2 := opPriority[t2.Value()]
	if ok1 && ok2 {
		return p1 > p2, nil
	}
	return false, errors.New("comparePriority: undefined token")
}
