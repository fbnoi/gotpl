package main

import (
	"gotpl/template"
	"strings"
)

// s           -> addExpress | boolExpr
// addExpr     -> mulExpr {opt1 mulExpr}
// boolExpr    -> mulExpr {opt3|opt4 mulExpr}
// mulExpr     -> arg {opt2|opt5 arg}
// arg         -> variable | automic | method
// method      -> name{.name}({(variable|arg)}{,(variable|arg)})
// variable    -> name{.name}
// automic     -> number|string
// opt1        -> + | -
// opt2        -> * | / | %
// opt3        -> > | < | >= | <=
// opt4        -> and | or | not
// opt5        -> []

const (
	TypeS = iota + 1
	TypeAddExpr
	TypeBoolExpr
	TypeMulExpr
	TypeArg
	TypeMethod
	TypeVariable
	TypeAtomic
	TypeOpt1
	TypeOpt2
	TypeOpt3
	TypeOpt4
	TypeOpt5
)

type Node struct {
	Type      int
	Signature strings.Builder
	Pn        *Node
	Ln        *Node
	Rn        *Node
	params    []*Node
}

type Lex struct {
	Stream  *template.TokenStream
	Current *Node
	AST     *Node
}

func (l *Lex) S() *Node {
	for !l.Stream.IsEOF() {
		token := l.Stream.Next()
		switch token.Type() {
		case template.TYPE_NAME:
			if l.Current.Type == TypeVariable {
				l.Current.Signature.WriteString(token.Value())
			} else {
				panic("")
			}
		case template.TYPE_STRING:
			if l.Current.Type == TypeMethod {
				sb := strings.Builder{}
				sb.WriteString(token.Value())
				l.Current.params = append(l.Current.params, &Node{Signature: sb, Type: TypeArg})
			} else {
				panic("")
			}
		case template.TYPE_NUMBER:
			switch l.Current.Type {
			case TypeMethod:
				sb := strings.Builder{}
				sb.WriteString(token.Value())
				node := &Node{Signature: sb, Type: TypeArg}
				l.Current.params = append(l.Current.params, node)
				l.Current = node
			}
		}
	}
	return l.AST
}
