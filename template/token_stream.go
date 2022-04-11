package template

import (
	"bytes"
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

func (ts *TokenStream) Next() *Token {
	ts.current++
	if ts.current >= len(ts.tokens) {
		panic("Unexpected end of template")
	}
	return ts.tokens[ts.current-1]
}

func (ts *TokenStream) IsEOF() bool {
	return TYPE_EOF == ts.tokens[ts.current].Type()
}

func (ts *TokenStream) GetSource() *Source {
	return ts.source
}
