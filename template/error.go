package template

import "fmt"

type UnexpectedEndOfFile struct {
	Source  *Source
	Line    int
	Message string
}
type UnexpectedToken struct {
	Source  *Source
	Line    int
	Message string
}
type ParseTemplateFaild struct {
	Source  *Source
	Line    int
	Message string
}

func (e *UnexpectedEndOfFile) Error() string { return e.Message }
func (e *UnexpectedToken) Error() string     { return e.Message }
func (e *ParseTemplateFaild) Error() string  { return e.Message }

func (e *UnexpectedEndOfFile) Overview() []*Line { return e.Source.Overview(e.Line) }
func (e *UnexpectedToken) Overview() []*Line     { return e.Source.Overview(e.Line) }
func (e *ParseTemplateFaild) Overview() []*Line  { return e.Source.Overview(e.Line) }

func NewUnexpectedEndOfFile(src *Source, line int, tok string) *UnexpectedEndOfFile {
	return &UnexpectedEndOfFile{
		Source:  src,
		Line:    line,
		Message: fmt.Sprintf("unexpected EOF at line: %s", line),
	}
}

func NewUnexpectedToken(src *Source, line int, tok string) *UnexpectedToken {
	return &UnexpectedToken{
		Source:  src,
		Line:    line,
		Message: fmt.Sprintf("unexpected token at line: %s", line),
	}
}

func NewParseTemplateFaild(src *Source, line int) *ParseTemplateFaild {
	return &ParseTemplateFaild{
		Source:  src,
		Line:    line,
		Message: "parse template failed",
	}
}
