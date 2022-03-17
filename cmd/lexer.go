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
			ts := &TextStmt{&BasicLit{ValuePos: template.Pos(token.At), Kind: template.TYPE_STRING, Value: token.Value()}}
			if filter.cursor == nil {
				filter.tr.List = append(filter.tr.List, ts)
			} else if st, ok := filter.cursor.(*IfStmt); ok {
				st.Body.List = append(st.Body.List, ts)
			} else if st, ok := filter.cursor.(*RangeStmt); ok {
				st.Body.List = append(st.Body.List, ts)
			} else if st, ok := filter.cursor.(*ForStmt); ok {
				st.Body.List = append(st.Body.List, ts)
			} else {
				panic("")
			}
		case template.TYPE_VAR_START:
			vs := &ValueStmt{}
			filter.internel = vs
			filter.internelExpr()
			if filter.cursor == nil {
				filter.tr.List = append(filter.tr.List, vs)
			} else if st, ok := filter.cursor.(*IfStmt); ok {
				st.Body.List = append(st.Body.List, vs)
			} else if st, ok := filter.cursor.(*RangeStmt); ok {
				st.Body.List = append(st.Body.List, vs)
			} else if st, ok := filter.cursor.(*ForStmt); ok {
				st.Body.List = append(st.Body.List, vs)
			} else {
				panic("")
			}
		case template.TYPE_BLOCK_START:
			token := stream.Next()
			switch token.Value() {
			case "if":
				is := &IfStmt{If: template.Pos(token.At)}
				filter.internel = is
				filter.internelExpr()
				if filter.cursor == nil {
					filter.tr.List = append(filter.tr.List, is)
				} else if st, ok := filter.cursor.(*IfStmt); ok {
					st.Body.List = append(st.Body.List, is)
				} else if st, ok := filter.cursor.(*RangeStmt); ok {
					st.Body.List = append(st.Body.List, is)
				} else if st, ok := filter.cursor.(*ForStmt); ok {
					st.Body.List = append(st.Body.List, is)
				} else {
					panic("")
				}
				if filter.cursor != nil {
					filter.stack = append(filter.stack, filter.cursor)
				}
				filter.cursor = is
			case "else":
				es := &SectionStmt{}
				filter.internelExpr()
				if st, ok := filter.cursor.(*IfStmt); ok {
					st.Body.List = append(st.Body.List, es)
				} else {
					panic("")
				}
				filter.cursor = es
			case "elseif":
				efs := &IfStmt{}
				filter.internelExpr()
				if st, ok := filter.cursor.(*IfStmt); ok {
					st.Body.List = append(st.Body.List, efs)
				} else {
					panic("")
				}
				filter.cursor = efs
			case "endif":
				filter.popIf()
			case "for":
				fs := &ForStmt{For: template.Pos(token.At)}
				filter.internel = fs
				filter.internelExpr()
				if filter.cursor == nil {
					filter.tr.List = append(filter.tr.List, fs)
				} else if st, ok := filter.cursor.(*IfStmt); ok {
					st.Body.List = append(st.Body.List, fs)
				} else if st, ok := filter.cursor.(*RangeStmt); ok {
					st.Body.List = append(st.Body.List, fs)
				} else if st, ok := filter.cursor.(*ForStmt); ok {
					st.Body.List = append(st.Body.List, fs)
				} else {
					panic("")
				}
				filter.cursor = fs
				filter.stack = append(filter.stack, fs)
			case "endfor":
				filter.popFor()
			case "range":
				rs := &ForStmt{For: template.Pos(token.At)}
				filter.internel = rs
				filter.internelExpr()
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
				filter.cursor = filter.popRange()
			case "block":
				bs := &BlockStmt{}
				filter.internel = bs
				filter.internelExpr()
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
				filter.internelExpr()
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

func (filter *TokenFilter) popBlock() *BlockStmt {
	if n, ok := filter.cursor.(*BlockStmt); ok {
		filter.cursor = filter.pop()
		return n
	}
	panic("")
}

func (filter *TokenFilter) popRange() *RangeStmt {
	if n, ok := filter.cursor.(*RangeStmt); ok {
		filter.cursor = filter.pop()
		return n
	}
	panic("")
}

func (filter *TokenFilter) popFor() *ForStmt {
	if n, ok := filter.cursor.(*ForStmt); ok {
		filter.cursor = filter.pop()
		return n
	}
	panic("")
}

func (filter *TokenFilter) popIf() *IfStmt {
	if n, ok := filter.cursor.(*IfStmt); ok {
		filter.cursor = filter.pop()
		return n
	}
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

func (filter *TokenFilter) internelExpr() {
	var (
		eStack  []Expr
		opStack []Expr
	)
	for !filter.stream.IsEOF() {
		token := filter.stream.Next()
		switch token.Type() {
		case template.TYPE_STRING, template.TYPE_NUMBER:
			eStack = append(opStack, &BasicLit{ValuePos: template.Pos(token.At), Kind: token.Type(), Value: token.Value()})
		case template.TYPE_NAME:
		case template.TYPE_OPERATOR:
		}
	}
}
