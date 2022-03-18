package main

import (
	"gotpl/template"
)

type Tree struct {
	List []Node
}

type TokenFilter struct {
	tr       *Tree
	cursor   Stmt
	internel Stmt
	stream   *template.TokenStream
	stack    []Stmt
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
				rs := &ForStmt{For: template.Pos(token.At)}
				filter.internel = rs
				// filter.internelExpr()
				if filter.cursor == nil {
					filter.tr.List = append(filter.tr.List, rs)
				} else if st, ok := filter.cursor.(*IfStmt); ok {
					st.Body.List = append(st.Body.List, rs)
				} else if st, ok := filter.cursor.(*RangeStmt); ok {
					st.Body.List = append(st.Body.List, rs)
				} else if st, ok := filter.cursor.(*ForStmt); ok {
					st.Body.List = append(st.Body.List, rs)
				} else {
					panic("")
				}
				filter.cursor = rs
				filter.stack = append(filter.stack, rs)
			case "endrange":
				filter.popRange()
			case "block":
				bs := &BlockStmt{}
				filter.internel = bs
				// filter.internelExpr()
				if filter.cursor == nil {
					filter.tr.List = append(filter.tr.List, bs)
				} else if st, ok := filter.cursor.(*IfStmt); ok {
					st.Body.List = append(st.Body.List, bs)
				} else if st, ok := filter.cursor.(*RangeStmt); ok {
					st.Body.List = append(st.Body.List, bs)
				} else if st, ok := filter.cursor.(*ForStmt); ok {
					st.Body.List = append(st.Body.List, bs)
				} else {
					panic("")
				}
				filter.cursor = bs
				filter.stack = append(filter.stack, bs)
			case "endblock":
				filter.popBlock()
			case "set":
				ss := &AssignStmt{}
				filter.internel = ss
				// filter.internelExpr()
				if filter.cursor == nil {
					filter.tr.List = append(filter.tr.List, ss)
				} else if st, ok := filter.cursor.(*IfStmt); ok {
					st.Body.List = append(st.Body.List, ss)
				} else if st, ok := filter.cursor.(*RangeStmt); ok {
					st.Body.List = append(st.Body.List, ss)
				} else if st, ok := filter.cursor.(*ForStmt); ok {
					st.Body.List = append(st.Body.List, ss)
				} else {
					panic("")
				}
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
		}
		break
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
		}
		break
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
	filter.internel = efs
	var ts []*template.Token
	for !filter.stream.IsEOF() {
		if token := filter.stream.Next(); token.Type() != template.TYPE_VAR_END {
			ts = append(ts, token)
		}
		break
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
	filter.internel = fs
	for !filter.stream.IsEOF() {
		var ts []*template.Token
		token := filter.stream.Next()
		for token.Value() != ";" && token.Type() != template.TYPE_EOF {
			ts = append(ts, token)
		}
		filter.internelExpr(ts)
	}
	filter.append(fs)
	filter.cursor = fs
	filter.stack = append(filter.stack, fs)
}

func (filter *TokenFilter) parseRange(token *template.Token) {
	rs := &RangeStmt{For: template.Pos(token.At)}
	filter.internel = rs
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

	if filter.cursor == nil {
		filter.tr.List = append(filter.tr.List, rs)
	} else if st, ok := filter.cursor.(*IfStmt); ok {
		st.Body.List = append(st.Body.List, rs)
	} else if st, ok := filter.cursor.(*RangeStmt); ok {
		st.Body.List = append(st.Body.List, rs)
	} else if st, ok := filter.cursor.(*ForStmt); ok {
		st.Body.List = append(st.Body.List, rs)
	} else {
		panic("")
	}
	filter.cursor = rs
	filter.stack = append(filter.stack, rs)
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

func (filter *TokenFilter) internelExpr([]*template.Token) Expr {
	return nil
	// var (
	// 	eStack  []Expr
	// 	opStack []Expr
	// )
	// for !filter.stream.IsEOF() {
	// 	token := filter.stream.Next()
	// 	switch token.Type() {
	// 	case template.TYPE_STRING, template.TYPE_NUMBER:
	// 		eStack = append(opStack, &BasicLit{ValuePos: template.Pos(token.At), Kind: token.Type(), Value: token.Value()})
	// 	case template.TYPE_NAME:
	// 		p := filter.stream.Peek()
	// 	case template.TYPE_OPERATOR:
	// 	}
	// }
}
