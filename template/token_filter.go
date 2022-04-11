package template

import (
	"fmt"
)

var opPriority = map[string]int{
	"==": 0, ">=": 0, "<=": 0, ">": 0, "<": 0, "!=": 0,
	"+": 5, "%": 10,
	"-": 5, "[": 15,
	"*": 10,
	"/": 10,
}

type Tree struct {
	List []ASTNode
}

type TokenFilter struct {
	Tr     *Tree
	cursor Stmt
	stream *TokenStream
	stack  []Stmt
}

func (filter *TokenFilter) Filter(stream *TokenStream) *Tree {
	filter.stream = stream
	for !stream.IsEOF() {
		token := stream.Next()
		switch token.Type() {
		case TYPE_TEXT:
			filter.parseText(token)
		case TYPE_VAR_START:
			filter.parseVar(token)
		case TYPE_BLOCK_START:
			token := stream.Next()
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
				filter.parseBlock(stream.Next())
			case "endblock":
				filter.popBlock()
			case "set":
				filter.parseSet(token)
			}
		}
	}
	return filter.Tr
}

func (filter *TokenFilter) parseText(token *Token) {
	ts := &TextStmt{&BasicLit{ValuePos: Pos(token.at), Kind: TYPE_STRING, Value: token.Value()}}
	filter.append(ts)
}

func (filter *TokenFilter) parseVar(token *Token) {
	vs := &ValueStmt{}
	var ts []*Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != TYPE_VAR_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	vs.Tok = filter.internelExpr(ts)
	filter.append(vs)
}

func (filter *TokenFilter) parseIf(token *Token) {
	is := &IfStmt{If: Pos(token.at)}
	var ts []*Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	is.Cond = filter.internelExpr(ts)
	filter.append(is)
	filter.push(is)
}

func (filter *TokenFilter) parseElse(token *Token) {
	es := &SectionStmt{}
	if st, ok := filter.cursor.(*IfStmt); ok {
		st.Else = es
	} else {
		panic("")
	}
	filter.push(es)
}

func (filter *TokenFilter) parseElseIf(token *Token) {
	efs := &IfStmt{}
	var ts []*Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	efs.Cond = filter.internelExpr(ts)
	if st, ok := filter.cursor.(*IfStmt); ok {
		st.Else = efs
	} else {
		panic("")
	}
	filter.push(efs)
}

func (filter *TokenFilter) parseFor(token *Token) {
	fs := &ForStmt{For: Pos(token.at)}
	var tss [][]*Token
	for !filter.stream.IsEOF() {
		var ts []*Token
		token := filter.stream.Next()
		for token.Value() != ";" && token.Type() != TYPE_EOF {
			ts = append(ts, token)
		}
		tss = append(tss, ts)
	}
	if len(tss) == 3 {
		fs.Init, fs.Cond, fs.Post = filter.parseAssignStmt(tss[0]), filter.internelExpr(tss[1]), filter.parseAssignStmt(tss[2])
	} else if len(tss) == 1 {
		fs.Cond = filter.internelExpr(tss[0])
	} else {
		panic("")
	}
	filter.append(fs)
	filter.push(fs)
}

func (filter *TokenFilter) parseRange(token *Token) {
	rs := &RangeStmt{For: Pos(token.at)}
	keyToken := filter.stream.Next()
	rs.Key = &Ident{NamePos: Pos(keyToken.at), Name: keyToken.Value()}
	valueToken := filter.stream.Next()
	if valueToken.Value() == "," {
		valueToken = filter.stream.Next()
	} else if valueToken.Value() == "=" {
		valueToken = nil
	}
	if valueToken != nil && valueToken.Value() != "_" {
		rs.Value = &Ident{NamePos: Pos(valueToken.at), Name: valueToken.Value()}
	}

	filter.append(rs)
	filter.push(rs)
}

func (filter *TokenFilter) parseBlock(token *Token) {
	if token.Type() != TYPE_NAME {
		panic("")
	}
	bs := &BlockStmt{Name: token.Value()}
	filter.append(bs)
	filter.push(bs)
}

func (filter *TokenFilter) parseSet(token *Token) {
	var ts []*Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != TYPE_BLOCK_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	ss := filter.parseAssignStmt(ts)
	filter.append(ss)
}

func (filter *TokenFilter) parseAssignStmt(ts []*Token) *AssignStmt {
	switch {
	case len(ts) == 0:
		return nil
	case len(ts) >= 2:
		token := ts[0]
		if token.Type() != TYPE_NAME {
			panic("")
		}
		ss := &AssignStmt{Lh: &Ident{NamePos: Pos(token.at), Name: token.Value()}}
		tok := ts[1]
		ss.TokPos, ss.Tok = Pos(tok.at), tok.Value()
		if len(ts) == 2 && (tok.Value() == "++" || tok.Value() == "--") {
			return ss
		} else if tok.Value() == "-=" || tok.Value() == "+=" || tok.Value() == "=" {
			ss.Rh = filter.internelExpr(ts[2:])
			return ss
		} else {
			ss.Rh = filter.internelExpr(ts[2:])
			return ss
		}
	}
	panic("")
}

func (filter *TokenFilter) append(s Stmt) {
	if filter.cursor == nil {
		filter.Tr.List = append(filter.Tr.List, s)
	} else if st, ok := filter.cursor.(*IfStmt); ok {
		if st.Body == nil {
			st.Body = &SectionStmt{}
		}
		st.Body.List = append(st.Body.List, s)
	} else if st, ok := filter.cursor.(*RangeStmt); ok {
		if st.Body == nil {
			st.Body = &SectionStmt{}
		}
		st.Body.List = append(st.Body.List, s)
	} else if st, ok := filter.cursor.(*ForStmt); ok {
		if st.Body == nil {
			st.Body = &SectionStmt{}
		}
		st.Body.List = append(st.Body.List, s)
	} else {
		panic("")
	}
}

func (filter *TokenFilter) popBlock() {
	_, ok := filter.cursor.(*BlockStmt)
	for !ok {
		filter.cursor = filter.pop()
		_, ok = filter.cursor.(*BlockStmt)
	}
	filter.cursor = filter.pop()
}

func (filter *TokenFilter) popRange() {
	_, ok := filter.cursor.(*RangeStmt)
	for !ok {
		filter.cursor = filter.pop()
		_, ok = filter.cursor.(*RangeStmt)
	}
	filter.cursor = filter.pop()
}

func (filter *TokenFilter) popFor() {
	_, ok := filter.cursor.(*ForStmt)
	for !ok {
		filter.cursor = filter.pop()
		_, ok = filter.cursor.(*ForStmt)
	}
	filter.cursor = filter.pop()
}

func (filter *TokenFilter) popIf() {
	_, ok := filter.cursor.(*IfStmt)
	for !ok {
		filter.cursor = filter.pop()
		_, ok = filter.cursor.(*IfStmt)
	}
	filter.cursor = filter.pop()
}

func (filter *TokenFilter) pop() Stmt {
	if len(filter.stack) == 0 {
		if filter.cursor == nil {
			panic("")
		}
		return nil
	}
	n := filter.stack[len(filter.stack)-1]
	filter.stack = filter.stack[:len(filter.stack)-1]
	return n
}

func (filter *TokenFilter) push(s Stmt) {
	if filter.cursor != nil {
		filter.stack = append(filter.stack, filter.cursor)
	}
	filter.cursor = s
}

func (filter *TokenFilter) internelExpr(ts []*Token) Expr {
	wrapper := &ExprWraper{stream: ts}

	return wrapper.Wrap()
}

type ExprWraper struct {
	stream  []*Token
	eStack  []Expr
	opStack []*Token
}

func (ew *ExprWraper) Wrap() Expr {
	for i := 0; i < len(ew.stream); i++ {
		token := ew.stream[i]
		switch token.Type() {
		case TYPE_STRING, TYPE_NUMBER:
			ew.pushExpr(&BasicLit{ValuePos: Pos(token.at), Value: token.Value()})
		case TYPE_NAME:
			if i+1 < len(ew.stream) {
				p := ew.stream[i+1]
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
				panic(token.Value())
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
				panic("")
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

func (ew *ExprWraper) peekExpr() Expr {
	if len(ew.eStack) == 0 {
		panic("")
	}
	return ew.eStack[len(ew.eStack)-1]
}

func (ew *ExprWraper) pushExpr(e Expr) {
	ew.eStack = append(ew.eStack, e)
}

func (ew *ExprWraper) popExpr() Expr {
	if len(ew.eStack) == 0 {
		panic("")
	}
	t := ew.eStack[len(ew.eStack)-1]
	ew.eStack = ew.eStack[:len(ew.eStack)-1]
	return t
}

func (ew *ExprWraper) peekOp() *Token {
	if len(ew.opStack) == 0 {
		panic("")
	}
	return ew.opStack[len(ew.opStack)-1]
}

func (ew *ExprWraper) pushOp(op *Token) {
	ew.opStack = append(ew.opStack, op)
}

func (ew *ExprWraper) popOp() (t *Token) {
	if len(ew.opStack) == 0 {
		panic("")
	}
	t = ew.opStack[len(ew.opStack)-1]
	ew.opStack = ew.opStack[:len(ew.opStack)-1]
	return
}

func waperBinary(op *Token, x1, x2 Expr) Expr {
	switch op.Type() {
	case TYPE_NAME:
		fn := &Ident{NamePos: Pos(op.at), Name: op.Value()}
		return &CallExpr{Fun: fn, Lparen: Pos(op.at + 1)}
	case TYPE_OPERATOR:
		switch op.Value() {
		case "+", "-", "*", "/", "%", ">", "<", ">=", "<=", "!=", "==":
			return &BinaryExpr{X: x2, Op: OpLit{OpPos: Pos(op.at), Op: op.Value()}, Y: x1}
		default:
			panic(op.Value())
		}
	case TYPE_PUNCTUATION:
		switch op.Value() {
		case "[":
			return &IndexExpr{X: x1, Index: x2}
		case ",":
			if arg, ok := x1.(*ArgsExpr); ok {
				arg.List = append(arg.List, x2)
				return arg
			}
			return &ArgsExpr{List: []Expr{x1, x2}}
		default:
			panic("")
		}
	default:
		panic("")

	}
}

func comparePriority(t1, t2 *Token) bool {
	if t1.Type() == TYPE_NAME {
		return true
	}
	if t1.Value() == "(" || t1.Value() == "[" {
		return true
	}

	if t2.Type() == TYPE_NAME {
		return false
	}

	if t2.Value() == "(" || t2.Value() == "[" {
		return false
	}

	if t2.Value() == "," {
		return true
	}

	p1, ok1 := opPriority[t1.Value()]
	p2, ok2 := opPriority[t2.Value()]
	if ok1 && ok2 {
		return p1 > p2
	}
	panic("")
}
