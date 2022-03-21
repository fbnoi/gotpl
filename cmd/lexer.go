package main

import (
	"gotpl/template"
)

var opPriority = map[string]int{
	"+": 0, "%": 5,
	"-": 0, "[": 10,
	"*": 5,
	"/": 5,
}

type Tree struct {
	List []Node
}

type TokenFilter struct {
	tr     *Tree
	cursor Stmt
	stream *template.TokenStream
	stack  []Stmt
}

func (filter *TokenFilter) Filter(stream *template.TokenStream) *Tree {
	filter.stream = stream
	for !stream.IsEOF() {
		token := stream.Next()
		switch token.Type() {
		case template.TYPE_TEXT:
			filter.parseText(token)
		case template.TYPE_VAR_START:
			filter.parseVar(token)
		case template.TYPE_BLOCK_START:
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
	return filter.tr
}

func (filter *TokenFilter) parseText(token *template.Token) {
	ts := &TextStmt{&BasicLit{ValuePos: template.Pos(token.At), Kind: template.TYPE_STRING, Value: token.Value()}}
	filter.append(ts)
}

func (filter *TokenFilter) parseVar(token *template.Token) {
	vs := &ValueStmt{}
	var ts []*template.Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != template.TYPE_VAR_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	vs.Tok = filter.internelExpr(ts)
	filter.append(vs)
}

func (filter *TokenFilter) parseIf(token *template.Token) {
	is := &IfStmt{If: template.Pos(token.At)}
	var ts []*template.Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != template.TYPE_VAR_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	is.Cond = filter.internelExpr(ts)
	filter.append(is)
	filter.push(is)
}

func (filter *TokenFilter) parseElse(token *template.Token) {
	es := &SectionStmt{}
	if st, ok := filter.cursor.(*IfStmt); ok {
		st.Else = es
	} else {
		panic("")
	}
	filter.push(es)
}

func (filter *TokenFilter) parseElseIf(token *template.Token) {
	efs := &IfStmt{}
	var ts []*template.Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != template.TYPE_VAR_END {
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

func (filter *TokenFilter) parseFor(token *template.Token) {
	fs := &ForStmt{For: template.Pos(token.At)}
	var tss [][]*template.Token
	for !filter.stream.IsEOF() {
		var ts []*template.Token
		token := filter.stream.Next()
		for token.Value() != ";" && token.Type() != template.TYPE_EOF {
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

func (filter *TokenFilter) parseRange(token *template.Token) {
	rs := &RangeStmt{For: template.Pos(token.At)}
	keyToken := filter.stream.Next()
	rs.Key = &Ident{NamePos: template.Pos(keyToken.At), Name: keyToken.Value()}
	valueToken := filter.stream.Next()
	if valueToken.Value() == "," {
		valueToken = filter.stream.Next()
	} else if valueToken.Value() == "=" {
		valueToken = nil
	}
	if valueToken != nil && valueToken.Value() != "_" {
		rs.Value = &Ident{NamePos: template.Pos(valueToken.At), Name: valueToken.Value()}
	}

	filter.append(rs)
	filter.push(rs)
}

func (filter *TokenFilter) parseBlock(token *template.Token) {
	if token.Type() != template.TYPE_NAME {
		panic("")
	}
	bs := &BlockStmt{Name: token.Value()}
	filter.append(bs)
	filter.push(bs)
}

func (filter *TokenFilter) parseSet(token *template.Token) {
	var ts []*template.Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != template.TYPE_VAR_END {
			ts = append(ts, token)
		} else {
			break
		}
	}
	ss := filter.parseAssignStmt(ts)
	filter.append(ss)
}

func (filter *TokenFilter) parseAssignStmt(ts []*template.Token) *AssignStmt {
	switch {
	case len(ts) == 0:
		return nil
	case len(ts) >= 2:
		token := ts[0]
		if token.Type() != template.TYPE_NAME {
			panic("")
		}
		ss := &AssignStmt{Lh: &Ident{NamePos: template.Pos(token.At), Name: token.Value()}}
		tok := ts[1]
		ss.TokPos, ss.Tok = template.Pos(tok.At), tok.Value()
		if len(ts) == 2 && (tok.Value() == "++" || tok.Value() == "--") {
			return ss
		} else if tok.Value() == "-=" || tok.Value() == "+=" || tok.Value() == "=" {
			ss.Rh = filter.internelExpr(ts[2:])
			return ss
		}
	}
	panic("")
}

func (filter *TokenFilter) append(s Stmt) {
	if filter.cursor == nil {
		filter.tr.List = append(filter.tr.List, s)
	} else if st, ok := filter.cursor.(*IfStmt); ok {
		st.Body.List = append(st.Body.List, s)
	} else if st, ok := filter.cursor.(*RangeStmt); ok {
		st.Body.List = append(st.Body.List, s)
	} else if st, ok := filter.cursor.(*ForStmt); ok {
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
	panic("")
}

func (filter *TokenFilter) popRange() {
	_, ok := filter.cursor.(*RangeStmt)
	for !ok {
		filter.cursor = filter.pop()
		_, ok = filter.cursor.(*RangeStmt)
	}
	filter.cursor = filter.pop()
	panic("")
}

func (filter *TokenFilter) popFor() {
	_, ok := filter.cursor.(*ForStmt)
	for !ok {
		filter.cursor = filter.pop()
		_, ok = filter.cursor.(*ForStmt)
	}
	filter.cursor = filter.pop()
	panic("")
}

func (filter *TokenFilter) popIf() {
	_, ok := filter.cursor.(*IfStmt)
	for !ok {
		filter.cursor = filter.pop()
		_, ok = filter.cursor.(*IfStmt)
	}
	filter.cursor = filter.pop()
	panic("")
}

func (filter *TokenFilter) pop() Stmt {
	if len(filter.stack) == 0 {
		panic("")
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

func (filter *TokenFilter) internelExpr(ts []*template.Token) Expr {

	var (
		eStack  []Expr
		opStack []Expr
		l       = len(ts)
	)
	for i := 0; i < l; i++ {
		token := ts[i]
		switch token.Type() {
		case template.TYPE_STRING, template.TYPE_NUMBER:
			eStack = append(eStack, &BasicLit{ValuePos: template.Pos(token.At), Kind: token.Type(), Value: token.Value()})
		case template.TYPE_NAME:
			if i+1 < l {
				p := ts[i+1]
				if p.Value() == "(" {
					fun := &Ident{NamePos: template.Pos(token.At), Name: token.Value()}
					opStack = append(opStack, &CallExpr{Fun: fun, Lparen: template.Pos(token.At)})
					continue
				}
			}
			eStack = append(eStack, &Ident{NamePos: template.Pos(token.At), Name: token.Value()})
		case template.TYPE_OPERATOR:

		}
	}
	return nil
}

type ExprStack struct {
	eStack  []Expr
	opStack []Expr
}

func (es *ExprStack) Push(e Expr) {
	switch x := e.(type) {
	case *BasicLit:
		x.exprNode()
	}
}

func (es *ExprStack) pushEStack(e Expr) {
	es.eStack = append(es.eStack, e)
}

func (es *ExprStack) pushOpStack(e Expr) {
	es.opStack = append(es.opStack, e)
}

func (es *ExprStack) comparePriority(e Expr) bool {
	// if
	// switch op := e.(type) {
	// case *CallExpr:
	// 	return false
	// case *OpLit:
	// 	if p1, ok := opPriority[op.Op]; ok {
	// 		op2 := es.opStack[len(es.opStack)-1]

	// 	}
	// }
}
