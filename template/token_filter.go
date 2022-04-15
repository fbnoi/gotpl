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
	for !stream.IsEOF() {
		token := filter.Next()
		switch token.Type() {
		case TYPE_TEXT:
			filter.parseText(token)
		case TYPE_VAR_START:
			filter.parseVar(token)
		case TYPE_BLOCK_START:
			token := filter.Next()
			switch token.Value() {
			case "if":
				filter.parseIf(token)
			case "else":
				filter.parseElse(token)
			case "elseif":
				filter.parseElseIf(token)
			case "endif":
				filter.popIf()
			case "for":
				filter.parseFor(token)
			case "endfor":
				filter.popFor()
			case "range":
				filter.parseRange(token)
			case "endrange":
				filter.popRange()
			case "block":
				filter.parseBlock(token)
			case "endblock":
				filter.popBlock()
			case "set":
				filter.parseSet(token)
			case "include":
				filter.parseInclude(token)
			case "extend":
				filter.parseExtend(token)
			default:
				return nil, filter.unexpected(token)
			}
		}
	}
	return filter.Tr, nil
}

func (filter *TokenFilter) parseExtend(token *Token) error {
	if filter.Tr.Extend != nil {
		return filter.unexpected(token)
	}
	es := &ExtendStmt{
		Extend: Pos(token.at),
	}
	if token := filter.Next(); token.Type() == TYPE_STRING {
		es.Ident = &BasicLit{
			ValuePos: Pos(token.at),
			Kind:     TYPE_STRING,
			Value:    token.Value(),
		}
		filter.Tr.Extend = es
		return nil
	}
	return filter.unexpected(token)
}

func (filter *TokenFilter) parseInclude(token *Token) error {
	is := &IncludeStmt{
		Include: Pos(token.at),
	}
	if token := filter.Next(); token.Type() == TYPE_STRING {
		is.Ident = &BasicLit{
			ValuePos: Pos(token.at),
			Kind:     TYPE_STRING,
			Value:    token.Value(),
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
				if as := parseAssignStmt(ts); as != nil {
					is.Params = append(is.Params, as)
				}
			}
		} else if token.Type() != TYPE_BLOCK_END {
			return filter.unexpected(token)
		}
		filter.append(is)
		return nil
	}
	return filter.unexpected(token)
}

func (filter *TokenFilter) parseText(token *Token) {
	t := &TextStmt{&BasicLit{
		ValuePos: Pos(token.at),
		Kind:     TYPE_STRING,
		Value:    token.Value(),
	}}
	filter.append(t)
}

func (filter *TokenFilter) parseVar(token *Token) {
	vs := &ValueStmt{}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_VAR_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	vs.Tok = parseExpr(ts)
	filter.append(vs)
}

func (filter *TokenFilter) parseIf(token *Token) {
	is := &IfStmt{If: Pos(token.at)}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	is.Cond = parseExpr(ts)
	filter.append(is)
	filter.push(is)
}

func (filter *TokenFilter) parseElse(token *Token) error {
	es := &SectionStmt{}
	if st, ok := filter.Cursor.(*IfStmt); ok {
		st.Else = es
	} else {
		return filter.unexpected(token)
	}
	filter.push(es)
	return nil
}

func (filter *TokenFilter) parseElseIf(token *Token) error {
	efs := &IfStmt{}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	efs.Cond = parseExpr(ts)
	if st, ok := filter.Cursor.(*IfStmt); ok {
		st.Else = efs
	} else {
		return filter.unexpected(token)
	}
	// elseif do not in stack
	filter.Cursor = efs
	return nil
}

func (filter *TokenFilter) parseFor(token *Token) error {
	fs := &ForStmt{For: Pos(token.at)}
	var tss [][]*Token
	for !filter.IsEOF() && token.Type() != TYPE_BLOCK_END {
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
		fs.Init, fs.Cond, fs.Post = parseAssignStmt(tss[0]), parseExpr(tss[1]), parseAssignStmt(tss[2])
	} else if len(tss) == 1 {
		fs.Cond = parseExpr(tss[0])
	} else {
		return filter.unexpected(token)
	}
	filter.append(fs)
	filter.push(fs)
	return nil
}

func (filter *TokenFilter) parseRange(token *Token) {
	rs := &RangeStmt{For: Pos(token.at)}
	keyToken := filter.Next()
	rs.Key = &Ident{NamePos: Pos(keyToken.at), Name: keyToken.Value()}
	valueToken := filter.Next()
	if valueToken.Value() == "," {
		valueToken = filter.Next()
	} else if valueToken.Value() == "=" {
		valueToken = nil
	}
	if valueToken != nil && valueToken.Value() != "_" {
		rs.Value = &Ident{NamePos: Pos(valueToken.at), Name: valueToken.Value()}
	}

	filter.append(rs)
	filter.push(rs)
}

func (filter *TokenFilter) parseBlock(token *Token) error {
	bs := &BlockStmt{
		Block: Pos(token.at),
		Name:  Ident{NamePos: Pos(token.at)},
	}
	if token.Type() != TYPE_NAME {
		return filter.unexpected(token)
	}
	filter.append(bs)
	filter.push(bs)
	return nil
}

func (filter *TokenFilter) parseSet(token *Token) {
	ss := &SetStmt{Set: Pos(token.at)}
	var ts []*Token
	for !filter.IsEOF() {
		if token := filter.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	ss.Assign = parseAssignStmt(ts)
	filter.append(ss)
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
	return NewParseTemplateFaild(filter.Source, int(filter.Cursor.End().Position()))
}

func (filter *TokenFilter) popBlock() {
	_, ok := filter.Cursor.(*BlockStmt)
	for !ok {
		filter.Cursor = filter.pop()
		_, ok = filter.Cursor.(*BlockStmt)
	}
	filter.Cursor = filter.pop()
}

func (filter *TokenFilter) popRange() {
	_, ok := filter.Cursor.(*RangeStmt)
	for !ok {
		filter.Cursor = filter.pop()
		_, ok = filter.Cursor.(*RangeStmt)
	}
	filter.Cursor = filter.pop()
}

func (filter *TokenFilter) popFor() {
	_, ok := filter.Cursor.(*ForStmt)
	for !ok {
		filter.Cursor = filter.pop()
		_, ok = filter.Cursor.(*ForStmt)
	}
	filter.Cursor = filter.pop()
}

func (filter *TokenFilter) popIf() {
	_, ok := filter.Cursor.(*IfStmt)
	for !ok {
		filter.Cursor = filter.pop()
		_, ok = filter.Cursor.(*IfStmt)
	}
	filter.Cursor = filter.pop()
}

func (filter *TokenFilter) pop() Stmt {
	if len(filter.Stack) == 0 {
		if filter.Cursor == nil {
			panic("")
		}
		return nil
	}
	n := filter.Stack[len(filter.Stack)-1]
	filter.Stack = filter.Stack[:len(filter.Stack)-1]
	return n
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

func (ew *ExprWraper) Wrap(stream []*Token) Expr {
	for i := 0; i < len(stream); i++ {
		token := stream[i]
		switch token.Type() {
		case TYPE_STRING, TYPE_NUMBER:
			ew.pushExpr(&BasicLit{ValuePos: Pos(token.at), Value: token.Value()})
		case TYPE_NAME:
			if i+1 < len(stream) {
				p := stream[i+1]
				if p.Value() == "(" {
					ew.pushOp(token)
					continue
				}
			}
			ew.pushExpr(&Ident{NamePos: Pos(token.at), Name: token.Value()})
		case TYPE_OPERATOR:
			switch token.Value() {
			case "+", "-", "*", "/", "%", "==", ">=", "<=", ">", "<", "!=":
				if len(ew.opStack) == 0 || comparePriority(token, ew.peekOp()) {
					ew.pushOp(token)
				} else {
					ew.revert(token)
				}
			default:
				panic(token)
			}
		case TYPE_PUNCTUATION:
			var op *Token
			switch token.Value() {
			case ",":
				op = ew.peekOp()
				for op.Value() != "(" && op.Value() != "," {
					op = ew.popOp()
					ew.revert(op)
				}
				ew.pushOp(token)
			case "(", "[":
				ew.pushOp(token)
			case "]":
				for op = ew.popOp(); op.Value() != "["; op = ew.popOp() {
					ew.revert(op)
				}
				ew.revert(op)
			case ")":
				for op = ew.popOp(); op.Value() != "("; op = ew.popOp() {
					ew.revert(op)
				}
				if op = ew.peekOp(); op.Type() == TYPE_NAME {
					op = ew.popOp()
					ew.revert(op)
				}
			default:
				panic(token)
			}
		}
	}
	for len(ew.opStack) > 0 {
		ew.revert(ew.popOp())
	}
	expr := ew.popExpr()
	if len(ew.eStack) > 0 {
		fmt.Println(ew.eStack)
		panic(len(ew.eStack))
	}
	return expr
}

func (ew *ExprWraper) revert(op *Token) {
	if op.Type() == TYPE_NAME {
		fun := &Ident{NamePos: Pos(op.at), Name: op.Value()}
		call := &CallExpr{Fun: fun, Lparen: Pos(op.at + 1)}
		if _, ok := ew.peekExpr().(*ArgsExpr); ok {
			args := ew.popExpr().(*ArgsExpr)
			call.Args = args
			call.Rparen = args.End() + 1
		} else {
			call.Rparen = call.Lparen + 1
		}
		ew.pushExpr(call)
		return
	}
	ew.pushExpr(waperBinary(op, ew.popExpr(), ew.popExpr()))
}

func (ew *ExprWraper) peekExpr() (Expr, error) {
	if len(ew.eStack) == 0 {
		return nil, errors.New("peekExpr: Try to peek expression from an empty stack")
	}
	return ew.eStack[len(ew.eStack)-1], nil
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

func (ew *ExprWraper) peekOp() (*Token, error) {
	if len(ew.opStack) == 0 {
		return nil, errors.New("peekOp: Try to pop token from an empty stack")
	}
	return ew.opStack[len(ew.opStack)-1], nil
}

func (ew *ExprWraper) pushOp(op *Token) {
	ew.opStack = append(ew.opStack, op)
}

func (ew *ExprWraper) popOp() (*Token, error) {
	if len(ew.opStack) == 0 {
		return nil, errors.New("peekOp: Try to pop token from an empty stack")
		panic("Resolve Expr failed, try to pop empty oprator stack")
	}
	t := ew.opStack[len(ew.opStack)-1]
	ew.opStack = ew.opStack[:len(ew.opStack)-1]
	return t, nil
}

func waperBinary(op *Token, x1, x2 Expr) (Expr, error) {
	switch op.Type() {
	case TYPE_NAME:
		fn := &Ident{NamePos: Pos(op.at), Name: op.Value()}
		return &CallExpr{Fun: fn, Lparen: Pos(op.at + 1)}, nil
	case TYPE_OPERATOR:
		switch op.Value() {
		case "+", "-", "*", "/", "%", ">", "<", ">=", "<=", "!=", "==":
			return &BinaryExpr{X: x2, Op: OpLit{OpPos: Pos(op.at), Op: op.Value()}, Y: x1}, nil
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
	return nil, errors.New(fmt.Sprintf("waperBinary: unexpected token %s", op.Value()))
}

func parseAssignStmt(ts []*Token) (*AssignStmt, error) {
	switch {
	case len(ts) == 0:
		return nil, errors.New("parseAssignStmt: empty token list")
	case len(ts) >= 2:
		token := ts[0]
		if token.Type() != TYPE_NAME {
			return nil, errors.New(fmt.Sprintf("parseAssignStmt: unexpected token %s", token.Value()))
		}
		ss := &AssignStmt{Lh: &Ident{NamePos: Pos(token.at), Name: token.Value()}}
		tok := ts[1]
		ss.TokPos, ss.Tok = Pos(tok.at), tok.Value()
		if len(ts) == 2 && (tok.Value() == "++" || tok.Value() == "--") {
			return ss, nil
		} else if tok.Value() == "-=" || tok.Value() == "+=" || tok.Value() == "=" {
			ss.Rh = parseExpr(ts[2:])
			return ss, nil
		}
	}
	return nil, errors.New("parseAssignStmt: parse failed")
}

func parseExpr(ts []*Token) Expr {
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
