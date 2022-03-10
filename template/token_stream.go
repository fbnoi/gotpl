package template

import (
	"bytes"
	"fmt"
)

type TokenStream struct {
	tokens  []*Token
	source  *Source
	current int
}

func (ts *TokenStream) String() string {
	buf := &bytes.Buffer{}
	for _, t := range ts.tokens {
		buf.WriteString(t.Value())
	}
	return buf.String()
}

func (ts *TokenStream) InjectToken(tks []*Token) {
	ts.tokens = append(ts.tokens, tks...)
}

func (ts *TokenStream) Next() *Token {
	ts.current++
	if ts.current >= len(ts.tokens) {
		et := ts.tokens[ts.current-1]
		panic(fmt.Sprintf("Unexpected end of template at line %d", et.Line()))
	}
	return ts.tokens[ts.current-1]
}

func (ts *TokenStream) Peek(n int) *Token {
	if ts.current+n >= len(ts.tokens) {
		et := ts.tokens[ts.current+n-1]
		panic(fmt.Sprintf("Unexpected end of template at line %d", et.Line()))
	}
	return ts.tokens[ts.current+n]
}

func (ts *TokenStream) IsEOF() bool {
	return TYPE_EOF == ts.tokens[ts.current].Type()
}

func (ts *TokenStream) GetCurrent() int {
	return ts.current
}

func (ts *TokenStream) GetSource() *Source {
	return ts.source
}
