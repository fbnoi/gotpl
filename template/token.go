package template

import "fmt"

const (
	TYPE_EOF = iota - 1
	TYPE_TEXT
	TYPE_BLOCK_START
	TYPE_VAR_START
	TYPE_BLOCK_END
	TYPE_VAR_END
	TYPE_METHOD
	TYPE_NAME
	TYPE_NUMBER
	TYPE_STRING
	TYPE_OPERATOR
	TYPE_PUNCTUATION
)

type Token struct {
	value string
	typ   int
	line  int
}

func (t *Token) String() string {
	return fmt.Sprintf("%s(%s)(%d)", TypeToString(t.typ), t.value, t.line)
}

func (t *Token) Value() string {
	return t.value
}

func (t *Token) Type() int {
	return t.typ
}

func (t *Token) Line() int {
	return t.line
}

func TypeToString(typ int) (name string) {
	switch typ {
	case TYPE_EOF:
		name = "TYPE_EOF"
	case TYPE_TEXT:
		name = "TYPE_TEXT"
	case TYPE_BLOCK_START:
		name = "TYPE_BLOCK_START"
	case TYPE_VAR_START:
		name = "TYPE_VAR_START"
	case TYPE_BLOCK_END:
		name = "TYPE_BLOCK_END"
	case TYPE_VAR_END:
		name = "TYPE_VAR_END"
	case TYPE_METHOD:
		name = "TYPE_METHOD"
	case TYPE_NAME:
		name = "TYPE_NAME"
	case TYPE_NUMBER:
		name = "TYPE_NUMBER"
	case TYPE_STRING:
		name = "TYPE_STRING"
	case TYPE_OPERATOR:
		name = "TYPE_OPERATOR"
	case TYPE_PUNCTUATION:
		name = "TYPE_PUNCTUATION"
	default:
		panic(fmt.Sprintf("Token of type '%d' does not exist.", typ))
	}
	return
}

func TypeToEnglish(typ int) string {
	switch typ {
	case TYPE_EOF:
		return "end of template"
	case TYPE_TEXT:
		return "text"
	case TYPE_BLOCK_START:
		return "begin of statement block"
	case TYPE_VAR_START:
		return "begin of print statement"
	case TYPE_BLOCK_END:
		return "end of statement block"
	case TYPE_VAR_END:
		return "end of print statement"
	case TYPE_METHOD:
		return "method"
	case TYPE_NAME:
		return "name"
	case TYPE_NUMBER:
		return "number"
	case TYPE_STRING:
		return "string"
	case TYPE_OPERATOR:
		return "operator"
	case TYPE_PUNCTUATION:
		return "punctuation"
	default:
		panic(fmt.Sprintf("Token of type '%d' does not exist.", typ))
	}
}
